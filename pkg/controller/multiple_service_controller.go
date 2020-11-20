package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	wccv1 "github.com/k8s-multiple-service/pkg/api/multipleservice/v1"
	informers "github.com/k8s-multiple-service/pkg/client/informers/externalversions/multipleservice/v1"
	listers "github.com/k8s-multiple-service/pkg/client/listers/multipleservice/v1"
)

const (
	maxRetries = 15
)

// MSController contains all kinds of clients
type MSController struct {
	// fromKubeClientset operates resource in cluster
	fromKubeClientset kubernetes.Interface
	// toKubeClientset operates resource in cluster
	toKubeClientset kubernetes.Interface

	serviceExportLister       listers.ServiceExportLister
	serviceExportListerSynced cache.InformerSynced

	// from cluster
	fromInformerFactory k8sinformers.SharedInformerFactory
	// to cluster
	toInformerFactory k8sinformers.SharedInformerFactory

	queue workqueue.RateLimitingInterface

	cache *serviceExportCache

	controllerResyncPeriod time.Duration
}

// NewMSController returns new msController
func NewMSController(fromKubeClientset, toKubeClientset kubernetes.Interface,
	fromInformerFactory k8sinformers.SharedInformerFactory,
	toInformerFactory k8sinformers.SharedInformerFactory,
	serviceExportInformer informers.ServiceExportInformer,
	controllerResyncPeriod time.Duration,
) (*MSController, error) {

	ms := &MSController{
		fromKubeClientset:      fromKubeClientset,
		toKubeClientset:        toKubeClientset,
		fromInformerFactory:    fromInformerFactory,
		toInformerFactory:      toInformerFactory,
		queue:                  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ms"),
		cache:                  &serviceExportCache{namespaceMap: make(map[string]map[string]int)},
		controllerResyncPeriod: controllerResyncPeriod,
	}

	serviceExportInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ms.addServiceExport,
		UpdateFunc: func(old, cur interface{}) {
			ms.updateServiceExport(old, cur)
		},
		DeleteFunc: ms.deleteServiceExport,
	})
	ms.serviceExportLister = serviceExportInformer.Lister()
	ms.serviceExportListerSynced = serviceExportInformer.Informer().HasSynced

	// TODO: watch toInformerFactory service
	fromInformerFactory.Core().V1().Services().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ms.enqueueService,
		UpdateFunc: func(old, cur interface{}) {
			curSVC := cur.(*v1.Service)
			oldSVC := old.(*v1.Service)
			if !reflect.DeepEqual(oldSVC.Spec.Ports, curSVC.Spec.Ports) {
				ms.enqueueService(cur)
			}
		},
		DeleteFunc: ms.enqueueService,
	})

	// endpoint's name、namespaces are equal to service's
	fromInformerFactory.Core().V1().Endpoints().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ms.enqueueService,
		UpdateFunc: func(old, cur interface{}) {
			curEP := cur.(*v1.Endpoints)
			oldEP := old.(*v1.Endpoints)
			if !reflect.DeepEqual(curEP.Subsets, oldEP.Subsets) {
				ms.enqueueService(cur)
			}
		},
		DeleteFunc: ms.enqueueService,
	})

	err := ms.initCache()
	if err != nil {
		klog.Errorf("ms controller initCache err: %v", err)
		return nil, err
	}

	return ms, nil
}

// Run begins reconcile
func (ms *MSController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer ms.queue.ShutDown()

	klog.Infof("Starting ms controller")
	defer klog.Infof("Shutting down ms controller")

	if !cache.WaitForCacheSync(stopCh, ms.serviceExportListerSynced,
		ms.fromInformerFactory.Core().V1().Services().Informer().HasSynced,
		ms.fromInformerFactory.Core().V1().Endpoints().Informer().HasSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(ms.worker, ms.controllerResyncPeriod, stopCh)
	}

	go func() {
		defer utilruntime.HandleCrash()
	}()

	<-stopCh
}

func (ms *MSController) worker() {
	for ms.processNextWorkItem() {
	}
}

func (ms *MSController) processNextWorkItem() bool {
	eKey, quit := ms.queue.Get()
	if quit {
		return false
	}
	defer ms.queue.Done(eKey)

	err := ms.syncService(eKey.(string))

	ms.handleErr(err, eKey)

	return true
}

func (ms *MSController) handleErr(err error, key interface{}) {
	if err == nil {
		ms.queue.Forget(key)
		return
	}

	if ms.queue.NumRequeues(key) < maxRetries {
		klog.Infof("Error syncing endpoints for service %q, retrying. Error: %v", key, err)
		ms.queue.AddRateLimited(key)
		return
	}

	klog.Warningf("Dropping service %q out of the queue: %v", key, err)
	ms.queue.Forget(key)
	utilruntime.HandleError(err)
}

// obj could be an *v1.Service, or a DeletionFinalStateUnknown marker item.
func (ms *MSController) enqueueService(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %+v: %v", obj, err))
		return
	}

	// check service belong to export
	exist, err := ms.serviceBelongToExistExport(key)
	if err != nil {
		klog.Errorf("serviceBelongToExistExport err: %v, ke: %v", err, key)
		return
	}
	if exist {
		klog.Infof("enqueueService key: %v", obj)
		ms.queue.Add(key)
	}
}

func (ms *MSController) syncService(key string) error {
	klog.Infof("begin syncService: %v", key)
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("SplitMetaNamespaceKey err: %v, key: %v", err, key)
		return err
	}

	service, err := ms.fromInformerFactory.Core().V1().Services().Lister().Services(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		// service absence in store means watcher caught the deletion, ensure LB info is cleaned
		klog.Infof("serviceExport has been deleted %v. Attempting to cleanup service resources", key)
		err = ms.processServiceDeletion(name, namespace)
	case err != nil:
		klog.Infof("Unable to retrieve serviceExport %v from store: %v", key, err)
	default:
		err = ms.processServiceUpdate(service, key)
	}

	return err
}

func (ms *MSController) processServiceDeletion(name, namespace string) error {
	desName := name + namespace
	toNamespaces, exist := ms.cache.get(namespace)
	if !exist {
		// log warning
		klog.Warningf("namespace: %v, is not in cache", namespace)
		return nil
	}

	for toNamespace := range toNamespaces {
		// delete endpoints
		err := ms.toKubeClientset.CoreV1().Endpoints(toNamespace).Delete(context.TODO(), desName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Errorf("ms controller processServiceDeletion Endpoints Delete err: %v", err)
			return err
		}

		// delete service
		err = ms.toKubeClientset.CoreV1().Services(toNamespace).Delete(context.TODO(), desName, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			klog.Errorf("ms controller processServiceDeletion Services Delete err: %v", err)
			return err
		}

		klog.Infof("success syncService: %v, delete svc: %v in namespaces: %v", name+"/"+namespace, desName, toNamespace)
	}

	return nil
}

func (ms *MSController) processServiceUpdate(service *v1.Service, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("SplitMetaNamespaceKey err: %v, key: %v", err, key)
		return err
	}

	endpoints, err := ms.fromInformerFactory.Core().V1().Endpoints().Lister().Endpoints(namespace).Get(name)
	if err != nil {
		klog.Errorf("ms controller processServiceUpdate Endpoints Get from cluster err: %v", err)
		return err
	}

	toServiceName := name + namespace
	toNamespaces, exist := ms.cache.get(namespace)
	if !exist {
		// log warning
		klog.Warningf("namespace: %v, is not in cache", namespace)
		return nil
	}

	for toNamespace := range toNamespaces {
		// get namesapce, or create
		newNamespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: toNamespace,
			},
		}

		_, err = ms.toKubeClientset.CoreV1().Namespaces().Create(context.TODO(), newNamespace, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			klog.Errorf("ms controller Namespaces create err: %v", err)
			return err
		}

		// Services operation
		newService := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      toServiceName,
				Namespace: toNamespace,
			},
			Spec: v1.ServiceSpec{
				Ports: service.Spec.Ports,
			},
		}

		toService, err := ms.toKubeClientset.CoreV1().Services(toNamespace).Get(context.TODO(), toServiceName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				_, err := ms.toKubeClientset.CoreV1().Services(toNamespace).Create(context.TODO(), newService, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("ms controller Services create err: %v", err)
					return err
				}
			} else {
				klog.Errorf("ms controller Services Get err: %v", err)
				return err
			}
		}

		if toService != nil && err == nil {
			toService.Spec.Ports = service.Spec.Ports
			_, err := ms.toKubeClientset.CoreV1().Services(toNamespace).Update(context.TODO(), toService, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("ms controller Services Update err: %v", err)
				return err
			}
		}

		// Endpoints operation
		newEndpoints := &v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      toServiceName,
				Namespace: toNamespace,
			},
			Subsets: endpoints.Subsets,
		}

		toEndpoints, err := ms.toKubeClientset.CoreV1().Endpoints(toNamespace).Get(context.TODO(), toServiceName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				_, err := ms.toKubeClientset.CoreV1().Endpoints(toNamespace).Create(context.TODO(), newEndpoints, metav1.CreateOptions{})
				if err != nil {
					klog.Errorf("ms controller Endpoints create err: %v", err)
					return err
				}
			} else {
				klog.Errorf("ms controller Endpoints Get err: %v", err)
				return err
			}
		}

		if toEndpoints != nil && err == nil {
			toEndpoints.Subsets = endpoints.Subsets
			_, err := ms.toKubeClientset.CoreV1().Endpoints(toNamespace).Update(context.TODO(), toEndpoints, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("ms controller Endpoints Update err: %v", err)
				return err
			}
		}

		klog.Infof("success syncService: %v, create/update svc: %v in namespaces: %v", key, toServiceName, toNamespace)
	}

	return nil
}

func (ms *MSController) initCache() error {
	serviceExports, err := ms.serviceExportLister.ServiceExports(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("ms controller list ServiceExports err: %v", err)
		return err
	}

	for _, serviceExport := range serviceExports {
		if len(serviceExport.Spec.FromNamespaces) == 0 || serviceExport.Spec.ToCluster == nil {
			continue
			// TODO: check labels or annotation
		}
		for _, namespace := range serviceExport.Spec.FromNamespaces {
			ms.cache.set(namespace, serviceExport.Spec.ToCluster.Namespace)
		}
		// no need for enqueue service
	}

	return nil
}

func (ms *MSController) addServiceExport(obj interface{}) {
	klog.Infof("namespaceMap: %v", ms.cache.namespaceMap)
	serviceExport := obj.(*wccv1.ServiceExport)
	if len(serviceExport.Spec.FromNamespaces) == 0 || serviceExport.Spec.ToCluster == nil {
		klog.Warningf("ms controller ServiceExports is not exporting service: %v", serviceExport)
		return
		// TODO: check labels or annotation
	}
	servicesTemp := make([]*v1.Service, 0)

	for _, namespace := range serviceExport.Spec.FromNamespaces {
		services, err := ms.fromInformerFactory.Core().V1().Services().Lister().Services(namespace).List(labels.Everything())
		if err != nil {
			klog.Errorf("ms controller list Services err: %v", err)
			return
		}
		servicesTemp = append(servicesTemp, services...)

		ms.cache.set(namespace, serviceExport.Spec.ToCluster.Namespace)
	}
	for _, service := range servicesTemp {
		// add queue
		ms.enqueueService(service)
	}

}

func (ms *MSController) updateServiceExport(old, cur interface{}) {
	klog.Infof("namespaceMap: %v", ms.cache.namespaceMap)
	curServiceExport := cur.(*wccv1.ServiceExport)
	oldServiceExport := old.(*wccv1.ServiceExport)

	if curServiceExport.Spec.ToCluster == nil || len(curServiceExport.Spec.FromNamespaces) == 0 {
		ms.deleteServiceExport(oldServiceExport)
		klog.Warningf("ms controller update ServiceExports is not exporting service: %v", curServiceExport)
		return
	}

	// Totally reset
	if curServiceExport.Spec.ToCluster.Namespace != oldServiceExport.Spec.ToCluster.Namespace {
		ms.deleteServiceExport(oldServiceExport)
		ms.addServiceExport(curServiceExport)
		return
	}

	toNamespace := curServiceExport.Spec.ToCluster.Namespace

	curNamespaces := curServiceExport.Spec.FromNamespaces
	oldNamespaces := oldServiceExport.Spec.FromNamespaces

	namespacesToDelete := make(map[string]int, 0)
	namespacesToAdd := make(map[string]int, 0)

	for _, oldNamespace := range oldNamespaces {
		namespacesToDelete[oldNamespace] = 1
	}

	for _, curNamespace := range curNamespaces {
		if _, ok := namespacesToDelete[curNamespace]; ok {
			delete(namespacesToDelete, curNamespace)
		} else {
			namespacesToAdd[curNamespace] = 1
		}
	}

	klog.Infof("namespacesToDelete: %v", namespacesToDelete)
	klog.Infof("namespacesToAdd: %v", namespacesToAdd)

	// add
	servicesTemp := make([]*v1.Service, 0)
	for namespace := range namespacesToAdd {
		services, err := ms.fromInformerFactory.Core().V1().Services().Lister().Services(namespace).List(labels.Everything())
		if err != nil {
			klog.Errorf("ms controller list Services in update err: %v", err)
			return
		}
		servicesTemp = append(servicesTemp, services...)
		ms.cache.set(namespace, toNamespace)
	}
	for _, service := range servicesTemp {
		ms.enqueueService(service)
	}

	// delete
	for namespace := range namespacesToDelete {
		// delete service from  cache
		services, err := ms.fromInformerFactory.Core().V1().Services().Lister().Services(namespace).List(labels.Everything())
		if err != nil {
			klog.Errorf("ms controller list Services in delete err: %v", err)
			return
		}

		ms.cache.delete(namespace, toNamespace)

		if !ms.cache.toNamespaceExist(namespace, toNamespace) {
			for _, service := range services {
				err := ms.toKubeClientset.CoreV1().Services(toNamespace).Delete(context.TODO(), service.Name+service.Namespace, metav1.DeleteOptions{})
				if err != nil {
					klog.Errorf("ms controller delete Services err: %v", err)
					return
				}
			}
		}
	}
}

func (ms *MSController) deleteServiceExport(obj interface{}) {
	klog.Infof("namespaceMap: %v", ms.cache.namespaceMap)
	serviceExport := obj.(*wccv1.ServiceExport)
	if len(serviceExport.Spec.FromNamespaces) == 0 || serviceExport.Spec.ToCluster == nil {
		klog.Warningf("ms controller delete ServiceExports is not exporting service: %v", serviceExport)
		return
		// TODO check labels or annotation
	}
	toNamespace := serviceExport.Spec.ToCluster.Namespace

	for _, namespace := range serviceExport.Spec.FromNamespaces {
		// delete namespace from  cache
		services, err := ms.fromInformerFactory.Core().V1().Services().Lister().Services(namespace).List(labels.Everything())
		if err != nil {
			klog.Errorf("ms controller list Services in delete err: %v", err)
			return
		}
		ms.cache.delete(namespace, toNamespace)

		if !ms.cache.toNamespaceExist(namespace, toNamespace) {
			for _, service := range services {
				err := ms.toKubeClientset.CoreV1().Services(toNamespace).Delete(context.TODO(), service.Name+service.Namespace, metav1.DeleteOptions{})
				if err != nil {
					klog.Errorf("ms controller delete Services err: %v", err)
					return
				}
			}
		}
	}
}

func (ms *MSController) serviceBelongToExistExport(key string) (bool, error) {
	namespace, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Errorf("ms controller SplitMetaNamespaceKey err: %v, key: %v", err, key)
		return false, err
	}
	_, exist := ms.cache.get(namespace)
	return exist, nil
}

// svc name must be no more than 63 characters
