package main

import (
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	clientset      kubernetes.Interface
	depLister      appslisters.DeploymentLister
	depCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:      clientset,
		depLister:      depInformer.Lister(),
		depCacheSynced: depInformer.Informer().HasSynced,
		queue: workqueue.NewRateLimitingQueueWithConfig(workqueue.DefaultControllerRateLimiter(), workqueue.RateLimitingQueueConfig{
			Name: "dep",
		}),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Printf("starting controller\n")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
		fmt.Printf("waiting for cache to be synced\n")
	}

	// every second run worker????? any problem???
	// worker cocurrency ????
	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *controller) worker() {
	for c.processItem() {
	}
}

func (c *controller) processItem() bool {
	// check check check the  role of the queue
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Forget(item)

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("getting key from cache %s\n", err.Error())
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("getting ns and name from key %s\n", err.Error())
		return false
	}

	// check if the object has been deleted from k8s cluster
	_, err = c.depLister.Deployments(ns).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			fmt.Printf("handle delete event for dep %s\n", err.Error())
			// delete service
			return true
		}
	}

	err = c.syncDeployment(ns, name)
	if err != nil {
		fmt.Printf("syncing deployment %s\n", err.Error())
		return false
	}

	return true
}

func (c *controller) syncDeployment(ns, name string) error {
	fmt.Printf("syncing syncing syncing\n")
	return nil
}

func (c *controller) handleAdd(obj interface{}) {
	fmt.Printf("deployment is added\n")
	c.queue.Add(obj)
}

func (c *controller) handleDel(obj interface{}) {
	fmt.Printf("deployment is deleted\n")
	c.queue.Add(obj)
}
