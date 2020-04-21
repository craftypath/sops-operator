package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SopsSecretSpec defines the desired state of SopsSecret
type SopsSecretSpec struct {
	// StringData allows specifying Sops-encrypted secret data in string form.
	StringData map[string]string `json:"stringData,omitempty"`

	// Type specifies the type of the secret.
	Type v1.SecretType `json:"type,omitempty"`
}

// SopsSecretStatus defines the observed state of SopsSecret
type SopsSecretStatus struct {
	LastUpdate metav1.Time
	Reason     string
	Status     string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SopsSecret is the Schema for the sopssecrets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sopssecrets,scope=Namespaced
type SopsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SopsSecretSpec   `json:"spec,omitempty"`
	Status SopsSecretStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SopsSecretList contains a list of SopsSecret
type SopsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SopsSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SopsSecret{}, &SopsSecretList{})
}
