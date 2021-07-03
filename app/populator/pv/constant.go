package main

const (
	annotationPopulatedFrom      = "demo.io/populated-from"
	annotationSelectedNode       = "volume.kubernetes.io/selected-node"
	annotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

const (
	groupDemoIO = "demo.io"
	versionV1   = "v1"
)

const (
	pvpKind     = "PersistentVolumePopulator"
	pvpResource = "persistentvolumepopulators"
)
