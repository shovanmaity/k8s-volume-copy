package main

import (
	"flag"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var (
	pvpGK  = schema.GroupKind{Group: groupDemoIO, Kind: pvpKind}
	pvpGVR = schema.GroupVersionResource{Group: groupDemoIO, Version: versionV1, Resource: pvpResource}

	vrGK  = schema.GroupKind{Group: groupDemoIO, Kind: vrKind}
	vrGVR = schema.GroupVersionResource{Group: groupDemoIO, Version: versionV1, Resource: vrResource}
)

func main() {
	klog.InitFlags(nil)
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("error getting k8s config error: %s", err)
		}
	}
	runController(cfg)
}
