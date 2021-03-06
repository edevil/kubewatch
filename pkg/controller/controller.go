/*
Copyright 2016 Skippbox, Ltd.
Copyright 2017 André Cruz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edevil/kubewatch/config"
	"github.com/edevil/kubewatch/pkg/handlers"
	"github.com/edevil/kubewatch/pkg/utils"
	"github.com/kr/pretty"

	"k8s.io/apimachinery/pkg/api/meta"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	api_v1 "k8s.io/client-go/pkg/api/v1"
	v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

// Controller object
type Controller struct {
	clientset     kubernetes.Interface
	queue         workqueue.RateLimitingInterface
	informer      cache.SharedIndexInformer
	eventHandlers []handlers.Handler
}

// QueueMessage - type of stuff stored on event queue
type QueueMessage struct {
	oldObject interface{}
	newObject interface{}
}

// Start - starts the controller
func Start(conf *config.Config, eventHandlers []handlers.Handler) {
	var kubeClient kubernetes.Interface
	if conf.InCluster {
		kubeClient = utils.GetClient()
	} else {
		kubeClient = utils.GetClientOutOfCluster()
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	if conf.Resource.Pod {
		c := newControllerPod(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	if conf.Resource.Service {
		c := newControllerService(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	if conf.Resource.ReplicationController {
		c := newControllerRC(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	if conf.Resource.Deployment {
		c := newControllerDeployment(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	if conf.Resource.Job {
		c := newControllerJob(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	if conf.Resource.PersistentVolume {
		c := newControllerPV(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	if conf.Resource.Event {
		c := newControllerEvent(kubeClient, eventHandlers, conf.InitialList)
		go c.Run(stopCh)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

func newControllerEvent(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().Events(meta_v1.NamespaceAll).List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.CoreV1().Events(meta_v1.NamespaceAll).Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &api_v1.Event{}, initialList)
}

func newControllerPod(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().Pods(meta_v1.NamespaceAll).List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.CoreV1().Pods(meta_v1.NamespaceAll).Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &api_v1.Pod{}, initialList)
}

func newControllerService(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().Services(meta_v1.NamespaceAll).List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.CoreV1().Services(meta_v1.NamespaceAll).Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &api_v1.Service{}, initialList)
}

func newControllerRC(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().ReplicationControllers(meta_v1.NamespaceAll).List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.CoreV1().ReplicationControllers(meta_v1.NamespaceAll).Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &api_v1.ReplicationController{}, initialList)
}

func newControllerDeployment(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.AppsV1beta1().Deployments(meta_v1.NamespaceAll).List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.AppsV1beta1().Deployments(meta_v1.NamespaceAll).Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &v1beta1.Deployment{}, initialList)
}

func newControllerJob(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.BatchV1().Jobs(meta_v1.NamespaceAll).List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.BatchV1().Jobs(meta_v1.NamespaceAll).Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &batchv1.Job{}, initialList)
}

func newControllerPV(client kubernetes.Interface, eventHandlers []handlers.Handler, initialList bool) *Controller {
	listFunc := func(options meta_v1.ListOptions) (runtime.Object, error) {
		return client.CoreV1().PersistentVolumes().List(options)
	}
	watchFunc := func(options meta_v1.ListOptions) (watch.Interface, error) {
		return client.CoreV1().PersistentVolumes().Watch(options)
	}
	return newControllerGeneric(client, eventHandlers, listFunc, watchFunc, &api_v1.PersistentVolume{}, initialList)
}

func newControllerGeneric(client kubernetes.Interface, eventHandlers []handlers.Handler, listFunc cache.ListFunc, watchFunc cache.WatchFunc, objType runtime.Object, initialList bool) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	var rVersion string

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				res, err := listFunc(options)
				if !initialList {
					if err != nil {
						log.Fatal("Could not fetch list:", err)
					}
					listMetaInterface, err := meta.ListAccessor(res)
					if err != nil {
						log.Fatalf("Unable to understand list result %v: %v", res, err)
					}
					rVersion = listMetaInterface.GetResourceVersion()

					return &v1.List{}, nil
				}

				return res, err
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if !initialList && options.ResourceVersion == "" && rVersion != "" {
					options.ResourceVersion = rVersion
				}
				return watchFunc(options)
			},
		},
		objType,
		0, //Skip resync
		cache.Indexers{},
	)

	c := &Controller{
		clientset:     client,
		informer:      informer,
		queue:         queue,
		eventHandlers: eventHandlers,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			queue.Add(&QueueMessage{
				newObject: obj,
			})
		},
		UpdateFunc: func(old, new interface{}) {
			queue.Add(&QueueMessage{
				newObject: new,
				oldObject: old,
			})
		},
		DeleteFunc: func(obj interface{}) {
			queue.Add(&QueueMessage{
				oldObject: obj,
			})
		},
	})

	return c
}

// Run starts the kubewatch controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	log.Println("Starting kubewatch controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	log.Println("Kubewatch controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processItem(key.(*QueueMessage))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		log.Printf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		log.Printf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(msg *QueueMessage) error {
	log.Printf("Processing change")

	if msg.oldObject == nil {
		for _, eHandler := range c.eventHandlers {
			eHandler.ObjectCreated(msg.newObject)
		}
	} else if msg.newObject == nil {
		for _, eHandler := range c.eventHandlers {
			eHandler.ObjectDeleted(msg.oldObject)
		}
	} else {
		changes := pretty.Diff(msg.oldObject, msg.newObject)
		for _, eHandler := range c.eventHandlers {
			eHandler.ObjectUpdated(msg.oldObject, msg.newObject, changes)
		}
	}

	return nil
}
