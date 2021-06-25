// +kubebuilder:object:generate=true
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RsyncPopulator is a volume populator that helps
// to create a volume from any rsync source.
type RsyncPopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec contains details of rsync source/ rsync daemon. Rsync client will
	// use these information to get the data for the volume.
	Spec RsyncPopulatorSpec `json:"spec"`
}

// RsyncPopulatorSpec contains the information of rsync daemon.
type RsyncPopulatorSpec struct {
	// Username is used as credential to access rsync daemon by the client.
	Username string `json:"username"`
	// Password is used as credential to access rsync daemon by the client.
	Password string `json:"password"`
	// Path represent mount path of the volume which we want to sync by the clinet.
	Path string `json:"path"`
	// URL is rsync daemon url it can be dns can be ip:port. Client will use
	// it to connect and get the data from daemon.
	URL string `json:"url"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RsyncPopulatorList is a list of RsyncPopulator objects
type RsyncPopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of RsyncPopulators
	Items []RsyncPopulator `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeClaimPopulator is a volume populator that helps to rename a
// PVC by applying patch on the PV of older PVC with new PVC.
type VolumeClaimPopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec contains volume claim information. ie - PVC name
	Spec VolumeClaimPopulatorSpec `json:"spec"`
}

// VolumeClaimPopulatorSpec contains information on Volume Claim.
type VolumeClaimPopulatorSpec struct {
	// PVCName is the name of the pvc which we want to rename.
	PVCName string `json:"pvcName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeClaimPopulatorList is a list of VolumeClaimPopulator objects
type VolumeClaimPopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of VolumeClaimPopulators
	Items []VolumeClaimPopulator `json:"items" protobuf:"bytes,2,rep,name=items"`
}
