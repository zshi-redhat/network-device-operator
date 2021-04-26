/*
Copyright 2021.

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

// NetDeviceSpec defines the desired state of NetDevice
type NetDeviceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Device Driver model configuration
	DeviceType NetDeviceType `json:"deviceType,omitempty"`
	// Device feature configuration (e.g. ethtool -K)
	DeviceFeature NetDeviceFeature `json:"deviceFeature,omitempty"`
}

// NetDeviceType defines the desired state of NetDeviceType
type NetDeviceType struct {
	// Driver bind configuration (vfio-pci, kernel drivers )
	Driver string `json:"driver,omitempty"`
	// Driver model configuration (switchdev, legacy)
	DriverModel string `json:"driverModel,omitempty"`
	// Link model configuration (Ethernet, InfiniBand)
	LinkType string `json:"linkType,omitempty"`
}

// NetDeviceFeature defines the desired state of NetDeviceFeature
type NetDeviceFeature struct {
	// Device features (tx-checksumming:on, rx-checksumming:off etc)
	Features map[string]string `json:"features,omitempty"`
}

// NetDeviceStatus defines the observed state of NetDevice
type NetDeviceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NetDevice is the Schema for the netdevices API
type NetDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetDeviceSpec   `json:"spec,omitempty"`
	Status NetDeviceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetDeviceList contains a list of NetDevice
type NetDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetDevice{}, &NetDeviceList{})
}
