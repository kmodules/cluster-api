package cloud

import (
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"context"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"golang.org/x/crypto/ssh"
	"fmt"
	"github.com/golang/glog"
)


type GenericUpgradeManager struct {
	ctx     context.Context
	kc      kubernetes.Interface
	ssh SSHGetter
	cluster *clusterv1.Cluster
}


var _ UpgradeManager = &GenericUpgradeManager{}

func NewUpgradeManager(ctx context.Context,ssh SSHGetter,  kc kubernetes.Interface, cluster *clusterv1.Cluster) UpgradeManager {
	return &GenericUpgradeManager{ctx: ctx,ssh:ssh,  kc: kc, cluster: cluster }
}

func (upm *GenericUpgradeManager) ExecuteSSHCommand(command string, machine *clusterv1.Machine) (string, error) {
	node, err := upm.kc.CoreV1().Nodes().Get(machine.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	cfg, err := upm.ssh.GetSSHConfig(upm.cluster, node)
	if err != nil {
		return "", err
	}

	keySigner, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return "", err
	}
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(keySigner),
		},
	}
	return ExecuteTCPCommand(command, fmt.Sprintf("%v:%v", cfg.HostIP, cfg.HostPort), config)
}

func (upm *GenericUpgradeManager) UpdateMasterInplace(oldMachine *clusterv1.Machine, newMachine *clusterv1.Machine) error {
	if oldMachine.Spec.Versions.ControlPlane != newMachine.Spec.Versions.ControlPlane {
		// First pull off the latest kubeadm.
		cmd := "export KUBEADM_VERSION=$(curl -sSL https://dl.k8s.io/release/stable.txt); " +
			"curl -sSL https://dl.k8s.io/release/${KUBEADM_VERSION}/bin/linux/amd64/kubeadm | sudo tee /usr/bin/kubeadm > /dev/null; " +
			"sudo chmod a+rx /usr/bin/kubeadm"
		_, err := upm.ExecuteSSHCommand(cmd, newMachine)
		if err != nil {
			glog.Infof("remotesshcomand error: %v", err)
			return err
		}

		// TODO: We might want to upgrade kubeadm if the target control plane version is newer.
		// Upgrade control plan.
		cmd = fmt.Sprintf("sudo kubeadm upgrade apply %s -y", "v"+newMachine.Spec.Versions.ControlPlane)
		_, err = upm.ExecuteSSHCommand(cmd, newMachine)
		if err != nil {
			glog.Infof("remotesshcomand error: %v", err)
			return err
		}
	}

	// Upgrade kubelet.
	if oldMachine.Spec.Versions.Kubelet != newMachine.Spec.Versions.Kubelet {
		cmd := fmt.Sprintf("sudo kubectl drain %s --kubeconfig /etc/kubernetes/admin.conf --ignore-daemonsets", newMachine.Name)
		// The errors are intentionally ignored as master has static pods.
		upm.ExecuteSSHCommand(cmd, newMachine)
		// Upgrade kubelet to desired version.
		cmd = fmt.Sprintf("sudo apt-get install kubelet=%s", newMachine.Spec.Versions.Kubelet+"-00")
		_, err := upm.ExecuteSSHCommand(cmd, newMachine)
		if err != nil {
			glog.Infof("remotesshcomand error: %v", err)
			return err
		}
		cmd = fmt.Sprintf("sudo kubectl uncordon %s --kubeconfig /etc/kubernetes/admin.conf", newMachine.Name)
		_, err = upm.ExecuteSSHCommand(cmd, newMachine)
		if err != nil {
			glog.Infof("remotesshcomand error: %v", err)
			return err
		}
	}

	return nil
}