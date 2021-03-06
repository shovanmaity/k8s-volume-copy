package main

import (
	"flag"
	"os"

	populator_machinery "github.com/kubernetes-csi/lib-volume-populator/populator-machinery"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	internalv1 "github.com/shovanmaity/k8s-volume-copy/client/apis/demo.io/v1"
)

const (
	prefix     = "demo.io"
	mountPath  = "/data"
	devicePath = "/dev/block"
)

func main() {
	klog.InitFlags(nil)
	if err := flag.Set("logtostderr", "true"); err != nil {
		panic(err)
	}

	var (
		imageName string
	)
	flag.StringVar(&imageName, "image-name", "", "Image to use for populating")
	flag.Parse()

	namespace := os.Getenv("POD_NAMESPACE")
	const (
		groupName  = "demo.io"
		apiVersion = "v1"
		kind       = "RsyncPopulator"
		resource   = "rsyncpopulators"
	)
	var (
		gk  = schema.GroupKind{Group: groupName, Kind: kind}
		gvr = schema.GroupVersionResource{Group: groupName, Version: apiVersion, Resource: resource}
	)
	populator_machinery.RunController("", "", imageName,
		namespace, prefix, gk, gvr, mountPath, devicePath, getPopulatorCmds, getPopulatorArgs, getPopulatorEnvs)
}

func getPopulatorCmds(rawBlock bool, u *unstructured.Unstructured) ([]string, error) {
	populator := internalv1.RsyncPopulator{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u.UnstructuredContent(), &populator)
	if nil != err {
		return nil, err
	}
	args := []string{
		"rsync",
	}
	return args, nil
}

func getPopulatorArgs(rawBlock bool, u *unstructured.Unstructured) ([]string, error) {
	populator := internalv1.RsyncPopulator{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u.UnstructuredContent(), &populator)
	if nil != err {
		return nil, err
	}
	args := []string{
		"-rv",
		"rsync://" + populator.Spec.Username + "@" +
			populator.Spec.URL + populator.Spec.Path,
		mountPath,
	}
	return args, nil
}

func getPopulatorEnvs(u *unstructured.Unstructured) ([]corev1.EnvVar, error) {
	populator := internalv1.RsyncPopulator{}
	err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(u.UnstructuredContent(), &populator)
	if nil != err {
		return nil, err
	}
	return []corev1.EnvVar{
		{
			Name:  "RSYNC_PASSWORD",
			Value: populator.Spec.Password,
		},
	}, nil
}
