package main

const (
	groupDemoIO = "demo.io"
	versionV1   = "v1"
)

const (
	pvpKind     = "PersistentVolumePopulator"
	pvpResource = "persistentvolumepopulators"
)

const (
	vrKind     = "VolumeRename"
	vrResource = "volumerenames"
)

const (
	provisionerNameAnnotation = "volume.beta.kubernetes.io/storage-provisioner"
	nodeNameAnnotation        = "volume.kubernetes.io/selected-node"
)

const (
	nameLabel      = "demo.io/name"
	appLabel       = "demo.io/app"
	roleLabel      = "demo.io/role"
	createdByLabel = "demo.io/created-by"
	managedByLabel = "demo.io/managed-by"
)

const (
	componentName = "pv-populator"
)
