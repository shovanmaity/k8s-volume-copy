// +kubebuilder:object:generate=true
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
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
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of RsyncPopulators
	Items []RsyncPopulator `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// persistentVolumePopulator is a volume populator that helps to rename a
// PVC by applying patch on the PV of older PVC with new PVC.
type PersistentVolumePopulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec contains volume claim information. ie - PVC name
	Spec PersistentVolumePopulatorSpec `json:"spec"`
}

// PersistentVolumePopulatorSpec contains information of Volume Claim.
type PersistentVolumePopulatorSpec struct {
	// PVCName is the name of the pvc which we want to rename.
	PVCName string `json:"pvcName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PersistentVolumePopulatorList is a list of PersistentVolumePopulator objects
type PersistentVolumePopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of PersistentVolumePopulators
	Items []PersistentVolumePopulator `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeCopy contains the informaion on making a copy of a volume.
// Volume copy can be created using any storage class.
type VolumeCopy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec contains details of rsync source/ rsync daemon. Rsync client will
	// use these information to get the data for the volume.
	Spec   VolumeCopySpec   `json:"spec"`
	Status VolumeCopyStatus `json:"status"`
}

// VolumeCopySpec contains information of source and new pvc.
type VolumeCopySpec struct {
	// PVCName is the PVC that we want to copy
	PVCName string `json:"pvcName"`
	// SCName is the storageclass of new PVC
	SCName string `json:"scName"`
	// NewName is new PVC name
	NewName string `json:"newName"`
}

// VolumeCopyStatus contains status of volume copy
type VolumeCopyStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeCopyList is a list of VolumeCopies objects
type VolumeCoppyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of VolumeCopies
	Items []VolumeCopy `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeRename contains the informaion on renaming a PVC.
type VolumeRename struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec contains details of renaming a PVC .
	Spec   VolumeRenameSpec   `json:"spec"`
	Status VolumeRenameStatus `json:"status"`
}

// VolumeRenameSpec contains PVC name and updated name
type VolumeRenameSpec struct {
	// PVCName is name of the PVC that we want to rename
	PVCName string `json:"pvcName"`
	// NewName is the updated name of the PVC
	NewName string `json:"newName"`
}

// VolumeRenameStatus contains status of the rename process.
type VolumeRenameStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeRenameList is a list of VolumeRenames objects
type VolumeRenameList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of VolumeRenames
	Items []VolumeRename `json:"items" protobuf:"bytes,2,rep,name=items"`
}
