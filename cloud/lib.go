package cloud

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	providerConf "sigs.k8s.io/cluster-api/cloud/providerconfig"
	apierrors "sigs.k8s.io/cluster-api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
)

func GetProviderconfig(codecFactory *serializer.CodecFactory, providerConfig clusterv1.ProviderConfig) (*providerConf.ProviderConfig, error) {
	obj, gvk, err := codecFactory.UniversalDecoder().Decode(providerConfig.Value.Raw, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding failure: %v", err)
	}

	config, ok := obj.(*providerConf.ProviderConfig)
	if !ok {
		return nil, fmt.Errorf("failure to cast to gce; type: %v", gvk)
	}

	return config, nil
}

// If the GCEClient has a client for updating Machine objects, this will set
// the appropriate reason/message on the Machine.Status. If not, such as during
// cluster installation, it will operate as a no-op. It also returns the
// original error for convenience, so callers can do "return handleMachineError(...)".
func HandleMachineError(machineClient client.MachineInterface, machine *clusterv1.Machine, err *apierrors.MachineError) error {
	if machineClient != nil {
		reason := err.Reason
		message := err.Message
		machine.Status.ErrorReason = &reason
		machine.Status.ErrorMessage = &message
		machineClient.UpdateStatus(machine)
	}

	glog.Errorf("Machine error: %v", err.Message)
	return err
}

func ValidateMachine(machine *clusterv1.Machine) *apierrors.MachineError {
	if machine.Spec.Versions.Kubelet == "" {
		return apierrors.InvalidMachineConfiguration("spec.versions.kubelet can't be empty")
	}
	if machine.Spec.Versions.ContainerRuntime.Name != "docker" {
		return apierrors.InvalidMachineConfiguration("Only docker is supported")
	}
	if machine.Spec.Versions.ContainerRuntime.Version != "1.12.0" {
		return apierrors.InvalidMachineConfiguration("Only docker 1.12.0 is supported")
	}
	return nil
}
