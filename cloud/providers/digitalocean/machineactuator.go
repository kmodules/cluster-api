package digitalocean

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"

	"github.com/golang/glog"
	"sigs.k8s.io/cluster-api/cloud"
	apierrors "sigs.k8s.io/cluster-api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/util"
	core "k8s.io/api/core/v1"
	"reflect"
	"net"
)

const (
	ZoneAnnotationKey    = "zone"
	NameAnnotationKey    = "name"

	UIDLabelKey       = "machine-crd-uid"
	BootstrapLabelKey = "boostrap"
)
func (cm *ClusterManager) PreparedActuator(token string, machineClient client.MachineInterface) error {
	conn, err := NewConnector(context.TODO())
	if err != nil {
		return err
	}

	// Only applicable if it's running inside machine controller pod.
	if machineClient != nil {
		conn.ssh, err = cloud.LoadSSHKey()
		fmt.Println(err)
		if err != nil {
			return err
		}
	}else {
		conn.ssh, _ = cloud.GenerateSSHKey()
	}
	cm.kubeadmToken = token
	cm.machineClient = machineClient
	cm.conn = conn

	return nil
}

func (cm *ClusterManager) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	config, err := cloud.GetProviderconfig(cm.codecFactory, machine.Spec.ProviderConfig)
	if err != nil {
		return cloud.HandleMachineError(cm.machineClient, machine, apierrors.InvalidMachineConfiguration(
			"Cannot unmarshal providerConfig field: %v", err))
	}
	cm.config = config

	if verr := cloud.ValidateMachine(machine); verr != nil {
		fmt.Println("error on validating")
		return cloud.HandleMachineError(cm.machineClient, machine, verr)
	}

	//var metadata string
	if machine.Spec.Versions.Kubelet == "" {
		return errors.New("invalid master configuration: missing Machine.Spec.Versions.Kubelet")
	}

	// skip error here, cause if not found instance then we need to create one
	instance, _ := cm.conn.instanceIfExists(machine)

	if instance == nil {
		if util.IsMaster(machine) {
			found, _, err := cm.conn.getPublicKey()
			if err != nil {
				return err
			}
			if !found {
				_, err = cm.conn.importPublicKey(cluster)
				if err != nil {
					return err
				}
			}
			if found, _ := cm.conn.getTags(cluster.Name); !found {
				if err = cm.conn.createTags(cluster.Name); err != nil {
					return err
				}
			}
		}
		labels := map[string]string{
			UIDLabelKey: fmt.Sprintf("%v", machine.ObjectMeta.UID),
		}
		if cm.machineClient == nil {
			labels[BootstrapLabelKey] = "true"
		}
		//machine.Labels = labels
		droplet, err := cm.conn.CreateInstance(cluster, machine, cm.kubeadmToken, config)

		if err != nil {
			fmt.Println("error on creating instance")
			return cloud.HandleMachineError(cm.machineClient, machine, apierrors.CreateMachine(
				"error creating instance: %v", err))
		}

		if util.IsMaster(machine) {
			err = cm.conn.addInstanceToPool(context.TODO(), cluster.Name, droplet.ID)
			if err != nil {
				return cloud.HandleMachineError(cm.machineClient, machine, apierrors.CreateMachine(
					"error creating instance: %v", err))
			}
		}
		// If we have a machineClient, then annotate the machine so that we
		// remember exactly what VM we created for it.
		if cm.machineClient != nil {
			return cm.updateAnnotations(machine)
		}
	} else {
		glog.Infof("Skipped creating a VM that already exists.\n")
	}

	return nil
}

func (cm *ClusterManager) GetIP(machine *clusterv1.Machine) (string, error) {
	lbStatus, _, err := cm.conn.GetLoadBalancer(machine.ClusterName)
	if err != nil {
		return "", err
	}
	return lbStatus.Ingress[0].IP, nil
}
func (cm *ClusterManager) GetKubeConfig(master *clusterv1.Machine) (string, error) {
	return "", nil
}
func (cm *ClusterManager) CreateLoadbalancer(machine *clusterv1.Machine) error {
	config, err := cloud.GetProviderconfig(cm.codecFactory, machine.Spec.ProviderConfig)
	if err != nil {
		return cloud.HandleMachineError(cm.machineClient, machine, apierrors.InvalidMachineConfiguration(
			"Cannot unmarshal providerConfig field: %v", err))
	}
	_,  err = cm.conn.EnsureLoadBalancer(machine.ClusterName, config.Zone)
	return err
}

// Create and start the machine controller. The list of initial
// machines don't have to be reconciled as part of this function, but
// are provided in case the function wants to refer to them (and their
// ProviderConfigs) to know how to configure the machine controller.
// Not idempotent.
func (cm *ClusterManager) CreateMachineController( cluster *clusterv1.Cluster, initialMachines []*clusterv1.Machine) error {

	fmt.Println("creating ssh key secret")
	if err := cloud.CreateSSHKeySecret(cm.conn.ssh, nil); err != nil {
		return err
	}
	fmt.Println("creating rolebinding")
	if err := cloud.CreateExtApiServerRoleBinding(); err != nil {
		return err
	}
	fmt.Println("creating apiserver and controller")
	if err := cloud.CreateApiServerAndController(cm.kubeadmToken, UID); err != nil {
		return err
	}
	return nil
}
func (cm *ClusterManager) PostDelete(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error {
	return nil
}

func (cm *ClusterManager)  Delete(machine *clusterv1.Machine) error {
	instance, err := cm.conn.instanceIfExists(machine)
	if err != nil {
		return err
	}

	if instance == nil {
		glog.Infof("Skipped deleting a VM that is already deleted.\n")
		return nil
	}

	conn, err := NewConnector(context.TODO())
	if err != nil {
		return err
	}
	cm.conn = conn


	if verr := cloud.ValidateMachine(machine); verr != nil {
		return cloud.HandleMachineError(cm.machineClient, machine, verr)
	}


	err = cm.conn.deleteInstance(context.TODO(), instance.ID)
	if err != nil {
		return cloud.HandleMachineError(cm.machineClient, machine, apierrors.DeleteMachine(
			"error deleting GCE instance: %v", err))
	}

	if cm.machineClient != nil {
		// Remove the finalizer
		machine.ObjectMeta.Finalizers = util.Filter(machine.ObjectMeta.Finalizers, clusterv1.MachineFinalizer)
		_, err = cm.machineClient.Update(machine)
	}

	return err
}
// Update the machine to the provided definition.
func (cm *ClusterManager) Update(cluster *clusterv1.Cluster, goalMachine *clusterv1.Machine) error {
	// Before updating, do some basic validation of the object first.
	config, err := cloud.GetProviderconfig(cm.codecFactory, goalMachine.Spec.ProviderConfig)
	if err != nil {
		return cloud.HandleMachineError(cm.machineClient, goalMachine, apierrors.InvalidMachineConfiguration(
			"Cannot unmarshal providerConfig field: %v", err))
	}
	cm.config = config
	cm.conn.ssh, err = cloud.LoadSSHKey()
	if err != nil {
		return err
	}

	if verr := cloud.ValidateMachine(goalMachine); verr != nil {
		fmt.Println("error on validating")
		return cloud.HandleMachineError(cm.machineClient, goalMachine, verr)
	}

	sm := cloud.NewStatusManager(cm.machineClient, cm.scheme)
	status, err := sm.InstanceStatus(goalMachine)
	if err != nil {
		return err
	}

	currentMachine := (*clusterv1.Machine)(status)
	if currentMachine == nil {
		instance, err := cm.conn.instanceIfExists(goalMachine)
		if err != nil {
			return err
		}
		if instance != nil  {
			glog.Infof("Populating current state for boostrap machine %v", goalMachine.ObjectMeta.Name)
			return cm.updateAnnotations(goalMachine)
		} else {
			return fmt.Errorf("Cannot retrieve current state to update machine %v", goalMachine.ObjectMeta.Name)
		}
	}

	if !cm.requiresUpdate(currentMachine, goalMachine) {
		return nil
	}

	apiserverAddress := cloud.APIServerAddress(cluster)
	kc, err := cloud.NewAdminClient(apiserverAddress)
	if err != nil {
		return fmt.Errorf("no cluster admin client found, %v", err)
	}

	if util.IsMaster(currentMachine) {
		upm := cloud.NewUpgradeManager(cm.ctx, cm, kc, cluster)
		glog.Infof("Doing an in-place upgrade for master.\n")
		err = upm.UpdateMasterInplace(currentMachine, goalMachine)
		if err != nil {
			glog.Errorf("master inplace update failed: %v", err)
		}
	} else {
		glog.Infof("re-creating machine %s for update.", currentMachine.ObjectMeta.Name)
		err = cm.Delete(currentMachine)
		if err != nil {
			glog.Errorf("delete machine %s for update failed: %v", currentMachine.ObjectMeta.Name, err)
		} else {
			err = cm.Create(cluster, goalMachine)
			if err != nil {
				glog.Errorf("create machine %s for update failed: %v", goalMachine.ObjectMeta.Name, err)
			}
		}
	}
	if err != nil {
		return err
	}
	err = cm.updateInstanceStatus(goalMachine)
	return err
	return nil
}
// Checks if the machine currently exists.
func (cm *ClusterManager) Exists(machine *clusterv1.Machine) (bool, error) {
	i, err := cm.conn.instanceIfExists(machine)
	if err != nil {
		return false, nil
	}
	return (i != nil), err
}

func (cm *ClusterManager) updateAnnotations(machine *clusterv1.Machine) error {
	config, err := cloud.GetProviderconfig(cm.codecFactory, machine.Spec.ProviderConfig)
	name := machine.ObjectMeta.Name
	zone := config.Zone

	if err != nil {
		return cloud.HandleMachineError(cm.machineClient, machine,
			apierrors.InvalidMachineConfiguration("Cannot unmarshal providerConfig field: %v", err))
	}

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = make(map[string]string)
	}
	//machine.ObjectMeta.Annotations[ProjectAnnotationKey] = project
	machine.ObjectMeta.Annotations[ZoneAnnotationKey] = zone
	machine.ObjectMeta.Annotations[NameAnnotationKey] = name
	_, err = cm.machineClient.Update(machine)
	if err != nil {
		return err
	}
	err = cm.updateInstanceStatus(machine)
	return err
}

// Sets the status of the instance identified by the given machine to the given machine
func (cm *ClusterManager) updateInstanceStatus(machine *clusterv1.Machine) error {
	fmt.Println("updating instance status")
	sm := cloud.NewStatusManager(cm.machineClient, cm.scheme)
	status := sm.Initialize(machine)
	currentMachine, err := util.GetCurrentMachineIfExists(cm.machineClient, machine)
	if err != nil {
		return err
	}

	if currentMachine == nil {
		// The current status no longer exists because the matching CRD has been deleted.
		return fmt.Errorf("Machine has already been deleted. Cannot update current instance status for machine %v", machine.ObjectMeta.Name)
	}

	m, err := sm.SetMachineInstanceStatus(currentMachine, status)
	if err != nil {
		return err
	}

	_, err = cm.machineClient.Update(m)
	return err
}

// The two machines differ in a way that requires an update
func (cm *ClusterManager) requiresUpdate(a *clusterv1.Machine, b *clusterv1.Machine) bool {
	// Do not want status changes. Do want changes that impact machine provisioning
	return !reflect.DeepEqual(a.Spec.ObjectMeta, b.Spec.ObjectMeta) ||
		!reflect.DeepEqual(a.Spec.ProviderConfig, b.Spec.ProviderConfig) ||
		!reflect.DeepEqual(a.Spec.Roles, b.Spec.Roles) ||
		!reflect.DeepEqual(a.Spec.Versions, b.Spec.Versions) ||
		a.ObjectMeta.Name != b.ObjectMeta.Name ||
		a.ObjectMeta.UID != b.ObjectMeta.UID
}


func (cm *ClusterManager) GetSSHConfig(cluster *clusterv1.Cluster, node *core.Node) (*clusterv1.SSHConfig, error) {
	cfg := &clusterv1.SSHConfig{
		PrivateKey: []byte(cm.conn.ssh.PrivateKey),
		User:       "root",
		HostPort:   int32(22),
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == core.NodeExternalIP {
			cfg.HostIP = addr.Address
		}
	}
	if net.ParseIP(cfg.HostIP) == nil {
		return nil, fmt.Errorf("failed to detect external Ip for node %s of cluster %s", node.Name, cluster.Name)
	}
	return cfg, nil
}
