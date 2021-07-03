package main

const (
	groupDemoIO = "demo.io"
	versionV1   = "v1"
)

const (
	rsyncKind     = "RsyncPopulator"
	rsyncResource = "rsyncpopulators"
)

const (
	vcKind     = "VolumeCopy"
	vcResource = "volumecopies"
)

const (
	populatorFinalizer = "demo.io/populate-target-protection"
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
	componentName = "volume-copy"
	populatorName = "rsync-populator"
)

const (
	rsyncClinetImage = "quay.io/k8s-volume-copy/rsync-client:ci"
	rsyncServerImage = "quay.io/k8s-volume-copy/rsync-daemon:ci"
	rsyncClientPass  = "pass"
)
