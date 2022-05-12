/*
Copyright 2022 AlexsJones.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LXDVirtualMachineSpec defines the desired state of LXDVirtualMachine
type LXDVirtualMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image string `json:"image,omitempty"`
}

// LXDVirtualMachineStatus defines the observed state of LXDVirtualMachine
type LXDVirtualMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LXDVirtualMachine is the Schema for the lxdvirtualmachines API
type LXDVirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LXDVirtualMachineSpec   `json:"spec,omitempty"`
	Status LXDVirtualMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LXDVirtualMachineList contains a list of LXDVirtualMachine
type LXDVirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LXDVirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LXDVirtualMachine{}, &LXDVirtualMachineList{})
}
