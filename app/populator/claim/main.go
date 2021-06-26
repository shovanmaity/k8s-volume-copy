package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

const (
	prefix = "demo.io"
)

func main() {
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		panic(err)
	}
	flag.Parse()

	namespace := os.Getenv("POD_NAMESPACE")
	const (
		groupName  = "demo.io"
		apiVersion = "v1"
		kind       = "VolumeClaimPopulator"
		resource   = "volumeclaimpopulators"
	)
	var (
		gk  = schema.GroupKind{Group: groupName, Kind: kind}
		gvr = schema.GroupVersionResource{Group: groupName, Version: apiVersion, Resource: resource}
	)
	runController(namespace, prefix, gk, gvr)
}
