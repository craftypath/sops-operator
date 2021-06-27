/*
Copyright The SOPS Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SopsSecretObjectMeta defines metadata for generated Secrets.
type SopsSecretObjectMeta struct {
	// Annotations allows adding annotations to generated Secrets.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels allows adding labels to generated Secrets.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// SopsSecretSpec defines the desired state of SopsSecret.
type SopsSecretSpec struct {
	// Metadata allows adding labels and annotations to generated Secrets.
	// +optional
	Metadata SopsSecretObjectMeta `json:"metadata,omitempty"`

	// StringData allows specifying Sops-encrypted secret data in string form.
	// +optional
	StringData map[string]string `json:"stringData,omitempty"`

	// Type specifies the type of the secret.
	// +optional
	Type corev1.SecretType `json:"type,omitempty"`
}

// SopsSecretStatus defines the observed state of SopsSecret.
type SopsSecretStatus struct {
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
	Reason     string      `json:"reason,omitempty"`
	Status     string      `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SopsSecret is the Schema for the sopssecrets API
type SopsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SopsSecretSpec   `json:"spec,omitempty"`
	Status SopsSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SopsSecretList contains a list of SopsSecret
type SopsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SopsSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SopsSecret{}, &SopsSecretList{})
}
