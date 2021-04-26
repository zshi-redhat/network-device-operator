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

// NetDevicePoolSpec defines the desired state of NetDevicePool
type NetDevicePoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Select the nodes
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Select devices on nodes
	DeviceSelector DeviceSelector `json:"deviceSelector,omitempty"`
	// K8s extended resource name
	ResourceName string `json:"resourceName,omitempty"`
	// NetDevice configuration
	Device NetDeviceConfig `json:"netDevice,omitempty"`
}

// DeviceSelector defines the desired state of DeviceSelector
type DeviceSelector struct {
	Vendors  []string `json:"vendors,omitempty"`
	Devices  []string `json:"devices,omitempty"`
	PciAddrs []string `json:"pciAddrs,omitempty"`
}

// NetDeviceConfig defines the desired state of network device configs
type NetDeviceConfig struct {
	// Device Driver model configuration
	Mode DeviceMode `json:"deviceType,omitempty"`
	// Device feature configuration (e.g. ethtool -K)
	Feature DeviceFeature `json:"deviceFeature,omitempty"`
}

// NetDeviceType defines the desired state of NetDeviceType
type DeviceMode struct {
	// Driver bind configuration (vfio-pci, kernel drivers )
	Driver string `json:"driver,omitempty"`
	// Driver model configuration (switchdev, legacy)
	DriverModel string `json:"driverModel,omitempty"`
	// Driver profile configuration (DDP)
	DriverProfile string `json:"driverProfile,omitempty"`
	// Link model configuration (Ethernet, InfiniBand)
	LinkType string `json:"linkType,omitempty"`
}

// DeviceFeature defines the desired state of NetDeviceFeature
type DeviceFeature struct {
	// Device features (tx-checksumming:on, rx-checksumming:off etc)
	Features map[string]string `json:"features,omitempty"`
}

// NetDevicePoolStatus defines the observed state of NetDevicePool
type NetDevicePoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NetDevicePool is the Schema for the netdevicepools API
type NetDevicePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetDevicePoolSpec   `json:"spec,omitempty"`
	Status NetDevicePoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetDevicePoolList contains a list of NetDevicePool
type NetDevicePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetDevicePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetDevicePool{}, &NetDevicePoolList{})
}
