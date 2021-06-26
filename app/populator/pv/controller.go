package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	runtimeu "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	internalv1 "github.com/shovanmaity/kvm/client/apis/demo.io/v1"
)

const (
	populatedFromAnnoSuffix      = "populated-from"
	annotationSelectedNode       = "volume.kubernetes.io/selected-node"
	annotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

type controller struct {
	populatorNamespace      string
	populatedFromAnnotation string
	pvcLister               listers.PersistentVolumeClaimLister
	pvcSynced               cache.InformerSynced
	pvLister                listers.PersistentVolumeLister
	pvSynced                cache.InformerSynced
	populatorLister         dynamiclister.Lister
	populatorSynced         cache.InformerSynced
	kubeClient              *kubernetes.Clientset
	dynamicClient           dynamic.Interface
	workqueue               workqueue.RateLimitingInterface
	gk                      schema.GroupKind
}

func runController(populatorNamespace, prefix string, gk schema.GroupKind, gvr schema.GroupVersionResource) {
	klog.Infof("Starting populator controller for %s", gk)
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		close(stopCh)
		<-sigCh
		os.Exit(1) // second signal. Exit directly.
	}()
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to create config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if nil != err {
		klog.Fatalf("Failed to create kube client: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if nil != err {
		klog.Fatalf("Failed to create dynamic client: %v", err)
	}
	informerFactory := informers.NewSharedInformerFactory(kubeClient, time.Second*30)
	dynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, time.Second*30)

	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
	pvInformer := informerFactory.Core().V1().PersistentVolumes()
	populatorInformer := dynamicInformerFactory.ForResource(gvr).Informer()

	c := &controller{
		populatorNamespace:      populatorNamespace,
		populatedFromAnnotation: prefix + "/" + populatedFromAnnoSuffix,
		pvcLister:               pvcInformer.Lister(),
		pvcSynced:               pvInformer.Informer().HasSynced,
		pvLister:                pvInformer.Lister(),
		pvSynced:                pvInformer.Informer().HasSynced,
		populatorLister:         dynamiclister.New(populatorInformer.GetIndexer(), gvr),
		populatorSynced:         populatorInformer.HasSynced,
		workqueue:               workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		kubeClient:              kubeClient,
		dynamicClient:           dynamicClient,
		gk:                      gk,
	}

	pvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handlePVC,
		UpdateFunc: func(old, new interface{}) {
			newPvc := new.(*corev1.PersistentVolumeClaim)
			oldPvc := old.(*corev1.PersistentVolumeClaim)
			if newPvc.ResourceVersion == oldPvc.ResourceVersion {
				return
			}
			c.handlePVC(new)
		},
		DeleteFunc: c.handlePVC,
	})

	informerFactory.Start(stopCh)
	dynamicInformerFactory.Start(stopCh)

	if err = c.run(stopCh); nil != err {
		klog.Fatalf("Failed to run controller: %v", err)
	}
}

func (c *controller) handlePVC(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtimeu.HandleError(err)
	}
	c.workqueue.Add(key)
}

func (c *controller) run(stopCh <-chan struct{}) error {
	defer runtimeu.HandleCrash()
	defer c.workqueue.ShutDown()

	ok := cache.WaitForCacheSync(stopCh, c.pvcSynced, c.pvSynced, c.populatorSynced)
	if !ok {
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
			runtimeu.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		parts := strings.Split(key, "/")
		klog.V(2).Infof("Processing PVC with key `%s`", key)
		if len(parts) != 2 {
			runtimeu.HandleError(fmt.Errorf("invalid resource key: %s", key))
			return nil
		}
		if done, err := c.syncPvc(context.TODO(), key, parts[0], parts[1]); err != nil || !done {
			c.workqueue.AddRateLimited(key)
			if err != nil {
				return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
			}
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
			runtimeu.HandleError(err)
		}
	}
}

func (c *controller) syncPvc(ctx context.Context, key, namespace, name string) (done bool, err error) {
	// Ignore PVCs present in working namespace
	if c.populatorNamespace == namespace {
		klog.V(2).Infof("Ignoring pvc `%s` present in populator namespace `%s`.", name, namespace)
		return true, nil
	}
	// Get pvc details
	pvc, err := c.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		klog.V(2).Infof("Error getting pvc, error: `%s`.", err)
		if errors.IsNotFound(err) {
			runtimeu.HandleError(fmt.Errorf("pvc '%s' in work queue no longer exists", key))
			return done, nil
		}
		return false, err
	}

	// Ignore PVCs without a datasource
	dataSource := pvc.Spec.DataSource
	if dataSource == nil {
		klog.V(2).Infof("Ignoring pvc `%s` present in namespace `%s`, datasource not present.", name, namespace)
		return true, nil
	}

	// Ignore PVCs that aren't for this populator to handle
	if c.gk.Group != *dataSource.APIGroup || c.gk.Kind != dataSource.Kind || dataSource.Name == "" {
		klog.V(2).Infof("Ignoring pvc `%s` present in namespace `%s`, datasource mismathc.", name, namespace)
		return true, nil
	}
	// Get populator object
	unstructured, err := c.populatorLister.Namespace(pvc.Namespace).Get(dataSource.Name)
	if nil != err {
		klog.V(2).Infof("Error getting populator, error: `%s`.", err)
		if errors.IsNotFound(err) {
			runtimeu.HandleError(fmt.Errorf("populator '%s' in work queue no longer exists", key))
			return false, nil
		}
		return false, err
	}
	populator := internalv1.PersistentVolumePopulator{}
	if err := runtime.DefaultUnstructuredConverter.
		FromUnstructured(unstructured.UnstructuredContent(), &populator); err != nil {
		return false, err
	}
	// Get source PVC
	source, err := c.pvcLister.PersistentVolumeClaims(namespace).Get(populator.Spec.PVCName)
	if err != nil {
		klog.V(2).Infof("Error getting pvc, error: `%s`.", err)
		if errors.IsNotFound(err) {
			runtimeu.HandleError(fmt.Errorf("pvc '%s' in work queue no longer exists", key))
			return true, nil
		}
		return false, err
	}
	// If source and pvc name is same then skip
	if pvc.Name == populator.Spec.PVCName {
		return true, nil
	}
	// If storageclass is not present in pvc and storageclass is not matching then skip it
	if pvc.Spec.StorageClassName != nil && source.Spec.StorageClassName != nil {
		if *pvc.Spec.StorageClassName != *source.Spec.StorageClassName {
			return true, nil
		}
	}
	nodeName := source.Annotations[annotationSelectedNode]
	provisioner := source.ObjectMeta.Annotations[annotationStorageProvisioner]
	if nodeName == "" || provisioner == "" {
		return false, nil
	}
	if pvc.Annotations[annotationSelectedNode] == "" || pvc.Annotations[annotationStorageProvisioner] == "" {
		pvcClone := pvc.DeepCopy()
		pvcClone.ObjectMeta.Annotations[annotationSelectedNode] = nodeName
		pvcClone.ObjectMeta.Annotations[annotationStorageProvisioner] = provisioner
		if _, err := c.kubeClient.CoreV1().PersistentVolumeClaims(pvcClone.Namespace).
			Update(context.TODO(), pvcClone, metav1.UpdateOptions{}); err != nil {
			return false, err
		}
	}

	// If the PVC is unbound, we need to perform the population
	if pvc.Spec.VolumeName == "" {
		pv, err := c.pvLister.Get(source.Spec.VolumeName)
		if err != nil {
			klog.V(2).Infof("Error getting pv, error: `%s`.", err)
			if errors.IsNotFound(err) {
				runtimeu.HandleError(fmt.Errorf("pv '%s' in work queue no longer exists", key))
				return false, nil
			}
			return false, err
		}

		// Examine the claimref for the PV and see if it's bound to the correct PVC
		claimRef := pv.Spec.ClaimRef
		if claimRef.Name != pvc.Name || claimRef.Namespace != pvc.Namespace || claimRef.UID != pvc.UID {
			// Make new PV with strategic patch values to perform the PV rebind
			patchPv := corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pv.Name,
					Annotations: map[string]string{},
				},
				Spec: corev1.PersistentVolumeSpec{
					ClaimRef: &corev1.ObjectReference{
						Namespace:       pvc.Namespace,
						Name:            pvc.Name,
						UID:             pvc.UID,
						ResourceVersion: pvc.ResourceVersion,
					},
				},
			}
			patchPv.Annotations[c.populatedFromAnnotation] = pvc.Namespace + "/" + dataSource.Name
			data, err := json.Marshal(patchPv)
			if nil != err {
				return false, err
			}
			if _, err := c.kubeClient.CoreV1().PersistentVolumes().Patch(ctx, pv.Name,
				types.StrategicMergePatchType, data, metav1.PatchOptions{}); err != nil {
				return false, err
			}
			// Don't start cleaning up yet -- we need to bind controller to acknowledge the switch
			return false, nil
		}
	}
	// If PVC' still exists, delete it
	if source != nil {
		if corev1.ClaimLost != source.Status.Phase {
			return false, nil
		}
		if err := c.kubeClient.CoreV1().PersistentVolumeClaims(source.Namespace).
			Delete(ctx, source.Name, metav1.DeleteOptions{}); err != nil {
			return false, err
		}
	}
	return true, nil
}
