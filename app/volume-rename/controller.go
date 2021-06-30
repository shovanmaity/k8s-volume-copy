package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	internalv1 "github.com/shovanmaity/k8s-volume-copy/client/apis/demo.io/v1"
)

type controller struct {
	kubeClient    *kubernetes.Clientset
	dynamicClient dynamic.Interface
	vrLister      dynamiclister.Lister
	vrSynced      cache.InformerSynced
	workqueue     workqueue.RateLimitingInterface
}

func runController(cfg *rest.Config) {
	klog.Infof("Starting populator controller for %s", vrGVR)
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		close(stopCh)
		<-sigCh
		os.Exit(1) // second signal. Exit directly.
	}()

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if nil != err {
		klog.Fatalf("Failed to create kube client: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if nil != err {
		klog.Fatalf("Failed to create dynamic client: %v", err)
	}

	dynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Second*30)
	vrInformer := dynamicInformerFactory.ForResource(vrGVR).Informer()
	c := &controller{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		vrLister:      dynamiclister.New(vrInformer.GetIndexer(), vrGVR),
		vrSynced:      vrInformer.HasSynced,
		workqueue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	vrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleVolumeRename,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.handleVolumeRename(newObj)
		},
		DeleteFunc: c.handleVolumeRename,
	})

	dynamicInformerFactory.Start(stopCh)
	if err := c.run(stopCh); nil != err {
		klog.Fatalf("Failed to run controller: %v", err)
	}
}

func (c *controller) handleVolumeRename(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
	}
	c.workqueue.Add(key)
}

func (c *controller) run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.vrSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	go wait.Until(c.runWorker, time.Second, stopCh)
	<-stopCh
	return nil
}

func (c *controller) runWorker() {
	processNext := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
			return nil
		}
		if err := c.syncPopulator(context.TODO(), key, parts[0], parts[1]); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		return nil
	}

	for {
		obj, shutdown := c.workqueue.Get()
		if shutdown {
			return
		}
		if err := processNext(obj); err != nil {
			utilruntime.HandleError(err)
		}
	}
}

/*
if found and not created by the populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensurePVC(want bool, namespace string, pvc *corev1.PersistentVolumeClaim) error {
	found := true
	obj, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
		Get(context.TODO(), pvc.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && obj.GetLabels()[createdByLabel] != componentName {
		return fmt.Errorf("resource found but not created by this operator")
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}

	if want && !found {
		_, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
			Create(context.TODO(), pvc, metav1.CreateOptions{})
		return err
	}
	if want && !found {
		err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
			Delete(context.TODO(), pvc.Name, metav1.DeleteOptions{})
		return err
	}
	return nil
}

/*
if found and not created by the populator then return error
if want and found return nil
if !want and !found return nil
if want and !found -> create return error/nil
if !want and found -> delete return error/nil
*/
func (c *controller) ensurePopulator(want bool, namespace string, populator *unstructured.Unstructured) error {
	found := true
	obj, err := c.dynamicClient.Resource(pvpGVR).Namespace(namespace).
		Get(context.TODO(), populator.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && obj.GetLabels()[createdByLabel] != componentName {
		return fmt.Errorf("resource found but not created by this operator")
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.dynamicClient.Resource(pvpGVR).Namespace(namespace).
			Create(context.TODO(), populator, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.dynamicClient.Resource(pvpGVR).Namespace(namespace).
			Delete(context.TODO(), populator.GetName(), metav1.DeleteOptions{})
		return err
	}
	return nil
}

func (c *controller) syncPopulator(ctx context.Context, key, namespace, name string) error {
	unstruct, err := c.vrLister.Namespace(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("populator '%s' in work queue no longer exists", key))
			return nil
		}
		return fmt.Errorf("error getting volume rename error: %s", err)
	}
	volumeRename := internalv1.VolumeRename{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.UnstructuredContent(),
		&volumeRename); err != nil {
		return fmt.Errorf("error converting volume rename `%s` in `%s` namespace error: %s",
			unstruct.GetName(), unstruct.GetNamespace(), err)
	}
	if volumeRename.Status.State == internalv1.VolumeRenameStatusCompleted ||
		volumeRename.Status.State == internalv1.VolumeRenameStatusFailed {
		klog.V(2).Infof("skip reconcile volume rename `%s` in `%s` namespace as it is in stable state.",
			volumeRename.GetName(), volumeRename.GetNamespace())
		return nil
	}
	if volumeRename.Status.State == "" {
		clone := volumeRename.DeepCopy()
		clone.Status.State = internalv1.VolumeRenameStatusInProgress
		volumeRenameMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&clone)
		if err != nil {
			return fmt.Errorf("error converting volume rename `%s` in `%s` namespace error: %s",
				volumeRename.GetName(), volumeRename.GetNamespace(), err)
		}
		volumeRenameUnstruct := unstructured.Unstructured{
			Object: volumeRenameMap,
		}
		if _, err := c.dynamicClient.Resource(vrGVR).Namespace(namespace).
			Update(context.TODO(), &volumeRenameUnstruct, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating volume rename %s in %s namespace error: %s",
				volumeRename.GetName(), volumeRename.GetNamespace(), err)
		}
		return nil
	}
	tc, err := templateFromVolumeRename(volumeRename)
	if err != nil {
		return fmt.Errorf("error creating template config error: %s", err)
	}
	populatorTemplate := tc.getPersistentVolumePopulatorTemplate()
	populatorTemplateMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&populatorTemplate)
	if err != nil {
		return fmt.Errorf("error converting populator %s in %s namespace error: %s",
			populatorTemplate.GetName(), namespace, err)
	}
	populatorTemplateUnstruct := unstructured.Unstructured{
		Object: populatorTemplateMap,
	}
	if err := c.ensurePopulator(true, namespace, &populatorTemplateUnstruct); err != nil {
		return fmt.Errorf("error ensuring(true) populator %s in %s namespace error: %s",
			populatorTemplate.GetName(), namespace, err)
	}
	pvc, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
		Get(context.TODO(), tc.pvcName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if err := c.ensurePopulator(false, namespace, &populatorTemplateUnstruct); err != nil {
				return fmt.Errorf("error ensuring(false) populator %s in %s namespace error: %s",
					populatorTemplate.GetName(), namespace, err)
			}
			found := true
			if _, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
				Get(context.TODO(), tc.newName, metav1.GetOptions{}); err != nil && errors.IsNotFound(err) {
				found = false
			}
			// Update status
			clone := volumeRename.DeepCopy()
			if found {
				clone.Status.State = internalv1.VolumeRenameStatusCompleted
			} else {
				clone.Status.State = internalv1.VolumeRenameStatusFailed
			}
			volumeRenameMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(clone)
			if err != nil {
				return fmt.Errorf("error converting volume rename `%s` in `%s` namespace error: %s",
					volumeRename.GetName(), volumeRename.GetNamespace(), err)
			}
			volumeRenameUnstruct := unstructured.Unstructured{
				Object: volumeRenameMap,
			}
			if _, err := c.dynamicClient.Resource(vrGVR).Namespace(namespace).
				Update(context.TODO(), &volumeRenameUnstruct, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("error updating volume rename %s in %s namespace error: %s",
					volumeRename.GetName(), volumeRename.GetNamespace(), err)
			}
			return nil
		}
		return fmt.Errorf("error getting pvc %s in %s namespace error: %s", tc.pvcName, namespace, err)
	}
	pvcTemplate := tc.getPVCDashTemplate(*pvc)
	if err := c.ensurePVC(true, namespace, &pvcTemplate); err != nil {
		return fmt.Errorf("error ensuring pvc %s in %s namespace error: %s", tc.pvcName, namespace, err)
	}
	return nil
}
