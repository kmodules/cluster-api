package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	Provider    string `json:"provider"`
	MachineType string `json:"machineType"`
	Image       string `json:"image"`
	Zone        string `json:"zone"`
}
