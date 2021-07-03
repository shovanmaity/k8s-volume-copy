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
	storagev1 "k8s.io/api/storage/v1"
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
	klog.Infof("Starting populator controller for %s", vcGK)
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
	vrInformer := dynamicInformerFactory.ForResource(vcGVR).Informer()
	c := &controller{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		vrLister:      dynamiclister.New(vrInformer.GetIndexer(), vcGVR),
		vrSynced:      vrInformer.HasSynced,
		workqueue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	vrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleVolumeCopy,
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.handleVolumeCopy(newObj)
		},
		DeleteFunc: c.handleVolumeCopy,
	})

	dynamicInformerFactory.Start(stopCh)
	if err := c.run(stopCh); nil != err {
		klog.Fatalf("Failed to run controller: %v", err)
	}
}

func (c *controller) handleVolumeCopy(obj interface{}) {
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

func (c *controller) syncPopulator(ctx context.Context, key, namespace, name string) error {
	unstruct, err := c.vrLister.Namespace(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("volume copy '%s' in work queue no longer exists", key))
			return nil
		}
		return fmt.Errorf("error getting volume rename error: %s", err)
	}
	volumeCopy := internalv1.VolumeCopy{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstruct.UnstructuredContent(),
		&volumeCopy); err != nil {
		return fmt.Errorf("error converting volume rename `%s` in `%s` namespace error: %s",
			unstruct.GetName(), unstruct.GetNamespace(), err)
	}
	if volumeCopy.Status.State == internalv1.StatusCompleted ||
		volumeCopy.Status.State == internalv1.StatusFailed {
		return nil
	}
	if volumeCopy.Status.State == "" {
		clone := volumeCopy.DeepCopy()
		clone.Status.State = internalv1.StatusInProgress
		if err := c.updateVolumeCopy(clone); err != nil {
			return fmt.Errorf("error updating status of volume copy `%s` in `%s` namespace, error: %s",
				volumeCopy.GetName(), volumeCopy.GetNamespace(), err)
		}
		return nil
	}
	volumeCopyClone := volumeCopy.DeepCopy()
	tc, err := templateFromVolumeCopy(*volumeCopyClone)
	if err != nil {
		return fmt.Errorf("error creating template config error: %s", err)
	}
	tc.imageName = rsyncServerImage
	tc.rsyncPassword = rsyncClientPass
	oldPVC, err := c.kubeClient.CoreV1().PersistentVolumeClaims(volumeCopy.Spec.SourceNamespace).
		Get(context.TODO(), volumeCopy.Spec.SourcePVC, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting pvc `%s` in `%s` namespace error: %s",
			volumeCopy.Spec.DestinationPVC, namespace, err)
	}
	pvcTempalte := tc.getPVCDashTemplate(*oldPVC)
	if err := c.ensurePVC(true, namespace, &pvcTempalte); err != nil {
		return fmt.Errorf("error ensuring pvc(true) `%s` in `%s` namespace, error: %s",
			pvcTempalte.GetName(), namespace, err)
	}
	populatorTemplate := tc.getRsyncPopulatorTemplate()
	if err := c.ensurePopulator(true, namespace, &populatorTemplate); err != nil {
		return fmt.Errorf("error ensuring(true) populator `%s` in `%s` namespace, error: %s",
			populatorTemplate.GetName(), namespace, err)
	}
	newPVC, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
		Get(context.TODO(), volumeCopy.Spec.DestinationPVC, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting pvc `%s` in `%s` namespace error: %s",
			volumeCopy.Spec.DestinationPVC, namespace, err)
	}
	sc, err := c.kubeClient.StorageV1().StorageClasses().
		Get(context.TODO(), volumeCopyClone.Spec.DestinationSC, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting pvc `%s` in `%s` namespace error: %s",
			volumeCopy.Spec.DestinationPVC, namespace, err)
	}
	waitForFirstConsumer := false
	if sc.VolumeBindingMode != nil && storagev1.VolumeBindingWaitForFirstConsumer == *sc.VolumeBindingMode {
		waitForFirstConsumer = true
	}
	selectedNode := ""
	if newPVC.GetAnnotations() != nil {
		selectedNode = newPVC.GetAnnotations()[nodeNameAnnotation]
	}
	if selectedNode == "" && waitForFirstConsumer {
		if volumeCopy.Status.State != internalv1.StatusWaitingForConsumer {
			clone := volumeCopy.DeepCopy()
			clone.Status.State = internalv1.StatusWaitingForConsumer
			if err := c.updateVolumeCopy(clone); err != nil {
				return fmt.Errorf("error updating status of volume copy `%s` in `%s` namespace, error: %s",
					volumeCopy.GetName(), volumeCopy.GetNamespace(), err)
			}
		}
		return nil
	}
	want := false
	if finalizers := newPVC.GetFinalizers(); finalizers != nil {
		for _, f := range finalizers {
			if f == populatorFinalizer {
				want = true
				break
			}
		}
	}
	if want {
		if volumeCopy.Status.State != internalv1.StatusInProgress {
			clone := volumeCopy.DeepCopy()
			clone.Status.State = internalv1.StatusInProgress
			if err := c.updateVolumeCopy(clone); err != nil {
				return fmt.Errorf("error updating status of volume copy `%s` in `%s` namespace, error: %s",
					volumeCopy.GetName(), volumeCopy.GetNamespace(), err)
			}
		}
		cmTemplate := tc.getCmTemplate()
		if err := c.ensureConfigMap(true, namespace, &cmTemplate); err != nil {
			return fmt.Errorf("error ensuring(true) configmap `%s` in `%s` namespace, error: %s",
				cmTemplate.GetName(), namespace, err)
		}
		podTemplate := tc.getPodTemplate()
		if err := c.ensurePod(true, namespace, &podTemplate); err != nil {
			return fmt.Errorf("error ensuring(true) pod `%s` in `%s` namespace, error: %s",
				podTemplate.GetName(), namespace, err)
		}
		svcTemplate := tc.getSvcTemplate()
		if err := c.ensureService(true, namespace, &svcTemplate); err != nil {
			return fmt.Errorf("error ensuring(true) service `%s` in `%s` namespace, error: %s",
				svcTemplate.GetName(), namespace, err)
		}
		return nil
	}
	cmTemplate := tc.getCmTemplate()
	if err := c.ensureConfigMap(false, namespace, &cmTemplate); err != nil {
		return fmt.Errorf("error ensuring(false) configmap `%s` in `%s` namespace, error: %s",
			cmTemplate.GetName(), namespace, err)
	}
	podTemplate := tc.getPodTemplate()
	if err := c.ensurePod(false, namespace, &podTemplate); err != nil {
		return fmt.Errorf("error ensuring(false) pod `%s` in `%s` namespace, error: %s",
			podTemplate.GetName(), namespace, err)
	}
	svcTemplate := tc.getSvcTemplate()
	if err := c.ensureService(false, namespace, &svcTemplate); err != nil {
		return fmt.Errorf("error ensuring(false) service `%s` in `%s` namespace, error: %s",
			svcTemplate.GetName(), namespace, err)
	}
	if err := c.ensurePopulator(false, namespace, &populatorTemplate); err != nil {
		return fmt.Errorf("error ensuring(false) populator `%s` in `%s` namespace, error: %s",
			populatorTemplate.GetName(), namespace, err)
	}
	if volumeCopy.Status.State != internalv1.StatusCompleted {
		clone := volumeCopy.DeepCopy()
		clone.Status.State = internalv1.StatusCompleted
		if err := c.updateVolumeCopy(clone); err != nil {
			return fmt.Errorf("error updating status of volume copy `%s` in `%s` namespace, error: %s",
				volumeCopy.GetName(), volumeCopy.GetNamespace(), err)
		}
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
func (c *controller) ensurePod(want bool, namespace string, pod *corev1.Pod) error {
	podClone := pod.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().Pods(namespace).
		Get(context.TODO(), podClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("resource found but not created by this operator")
	}
	if want && found {
		if obj.Status.Phase == corev1.PodFailed || obj.Status.Phase == corev1.PodSucceeded {
			c.ensurePod(false, namespace, podClone)
		}
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.kubeClient.CoreV1().Pods(namespace).
			Create(context.TODO(), podClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().Pods(namespace).
			Delete(context.TODO(), podClone.Name, metav1.DeleteOptions{})
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
func (c *controller) ensureService(want bool, namespace string, svc *corev1.Service) error {
	svcClone := svc.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().Services(namespace).
		Get(context.TODO(), svcClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("resource found but not created by this operator")
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.kubeClient.CoreV1().Services(namespace).
			Create(context.TODO(), svcClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().Services(namespace).
			Delete(context.TODO(), svcClone.Name, metav1.DeleteOptions{})
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
func (c *controller) ensureConfigMap(want bool, namespace string, cm *corev1.ConfigMap) error {
	cmClone := cm.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().ConfigMaps(namespace).
		Get(context.TODO(), cmClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("resource found but not created by this operator")
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.kubeClient.CoreV1().ConfigMaps(namespace).
			Create(context.TODO(), cmClone, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.kubeClient.CoreV1().ConfigMaps(namespace).
			Delete(context.TODO(), cmClone.Name, metav1.DeleteOptions{})
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
func (c *controller) ensurePVC(want bool, namespace string, pvc *corev1.PersistentVolumeClaim) error {
	pvcClone := pvc.DeepCopy()
	found := true
	obj, err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
		Get(context.TODO(), pvcClone.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
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
			Create(context.TODO(), pvcClone, metav1.CreateOptions{})
		return err
	}
	if want && !found {
		err := c.kubeClient.CoreV1().PersistentVolumeClaims(namespace).
			Delete(context.TODO(), pvcClone.Name, metav1.DeleteOptions{})
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
func (c *controller) ensurePopulator(want bool, namespace string, populator *internalv1.RsyncPopulator) error {
	found := true
	populatorClone := populator.DeepCopy()
	populatorMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&populatorClone)
	if err != nil {
		return err
	}
	populatorUnstruct := &unstructured.Unstructured{
		Object: populatorMap,
	}
	obj, err := c.dynamicClient.Resource(rsyncGVR).Namespace(namespace).
		Get(context.TODO(), populatorClone.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			found = false
		} else {
			return err
		}
	}
	if found && (obj.GetLabels() == nil || obj.GetLabels()[createdByLabel] != componentName) {
		return fmt.Errorf("resource found but not created by this operator")
	}
	if want && found {
		return nil
	}
	if !want && !found {
		return nil
	}
	if want && !found {
		_, err := c.dynamicClient.Resource(rsyncGVR).Namespace(namespace).
			Create(context.TODO(), populatorUnstruct, metav1.CreateOptions{})
		return err
	}
	if !want && found {
		err := c.dynamicClient.Resource(rsyncGVR).Namespace(namespace).
			Delete(context.TODO(), populatorClone.GetName(), metav1.DeleteOptions{})
		return err
	}
	return nil
}

// updateVolumeCopy updates a volume copy object
func (c *controller) updateVolumeCopy(vc *internalv1.VolumeCopy) error {
	vcClone := vc.DeepCopy()
	vcMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(vcClone)
	if err != nil {
		return err
	}
	vcUnstruct := &unstructured.Unstructured{
		Object: vcMap,
	}
	_, err = c.dynamicClient.Resource(vcGVR).Namespace(vcClone.GetNamespace()).
		Update(context.TODO(), vcUnstruct, metav1.UpdateOptions{})
	return err
}
