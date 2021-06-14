module github.com/shovanmaity/kubernetes-volume-migration

go 1.14

require (
	github.com/kubernetes-csi/lib-volume-populator v0.0.0-20210427161538-98a19e9b7590
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.19.9
	k8s.io/apimachinery v0.21.1
)

replace github.com/kubernetes-csi/lib-volume-populator => /home/shovan/Desktop/go/src/github.com/kubernetes-csi/lib-volume-populator
