module github.com/shovanmaity/k8s-volume-copy

go 1.14

require (
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/kubernetes-csi/lib-volume-populator v0.0.0-20210427161538-98a19e9b7590
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.8.0
	sigs.k8s.io/controller-tools v0.5.0
)

replace github.com/kubernetes-csi/lib-volume-populator => /home/shovan/Desktop/go/src/github.com/kubernetes-csi/lib-volume-populator
