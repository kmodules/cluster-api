package cloud

import (
	"github.com/pkg/errors"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/controller/machine"
//	"sigs.k8s.io/cluster-api/gcp-deployer/deploy"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	core "k8s.io/api/core/v1"
)

var (
	ErrNotFound       = errors.New("node not found")
	ErrNotImplemented = errors.New("not implemented")
	ErrNoMasterNG     = errors.New("cluster has no master NodeGroup")
)


type Interface interface {
	machine.Actuator

	PreparedActuator(token string, mc client.MachineInterface) error
	GetIP(machine *clusterv1.Machine) (string, error)
	GetKubeConfig(master *clusterv1.Machine) (string, error)
	CreateLoadbalancer(machine *clusterv1.Machine) error

	// Create and start the machine controller. The list of initial
	// machines don't have to be reconciled as part of this function, but
	// are provided in case the function wants to refer to them (and their
	// ProviderConfigs) to know how to configure the machine controller.
	// Not idempotent.
	CreateMachineController(cluster *clusterv1.Cluster, initialMachines []*clusterv1.Machine) error
	PostDelete(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error
	//InitializeMachineDeployer(kubeadmToken string, machineClient client.MachineInterface)(*deploy.MachineDeployer, error)
	//	GetDefaultNodeSpec(cluster *api.Cluster, sku string) (api.NodeSpec, error)
	//	SetDefaults(in *api.Cluster) error
	//	Apply(in *api.Cluster, dryRun bool) ([]api.Action, error)
	//	IsValid(cluster *api.Cluster) (bool, error)
	// GetAdminClient() (kubernetes.Interface, error)

	// IsValid(cluster *api.Cluster) (bool, error)
	// Delete(req *proto.ClusterDeleteRequest) error
	// SetVersion(req *proto.ClusterReconfigureRequest) error
	// Scale(req *proto.ClusterReconfigureRequest) error
	// GetInstance(md *api.InstanceStatus) (*api.Instance, error)
}

type SSHGetter interface {
	GetSSHConfig(cluster *clusterv1.Cluster, node *core.Node) (*clusterv1.SSHConfig, error)
}

type UpgradeManager interface {
	UpdateMasterInplace(oldMachine *clusterv1.Machine, newMachine *clusterv1.Machine) error
}
