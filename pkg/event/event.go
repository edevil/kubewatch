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

package event

import (
	"fmt"
	"log"
	"reflect"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/apps/v1beta1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/tools/cache"
)

// Event represent an event got from k8s api server
// Events from different endpoints need to be casted to KubewatchEvent
// before being able to be handled by handler
type Event struct {
	Namespace string
	Kind      string
	Component string
	Host      string
	Reason    string
	Status    string
	Name      string
	Message   string
}

var m = map[string]string{
	"created": "Normal",
	"deleted": "Danger",
	"updated": "Warning",
}

// New create new KubewatchEvent
func New(obj interface{}, action string) Event {
	var namespace, kind, component, host, reason, status, name, message string

	if deletedObj, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		obj = deletedObj.Obj
		log.Println("Caught an object whose final state was unknown...")
	}

	if apiService, ok := obj.(*v1.Service); ok {
		namespace = apiService.ObjectMeta.Namespace
		name = apiService.Name
		kind = "service"
		component = string(apiService.Spec.Type)
		reason = action
		status = m[action]
	} else if apiPod, ok := obj.(*v1.Pod); ok {
		namespace = apiPod.ObjectMeta.Namespace
		name = apiPod.Name
		kind = "pod"
		reason = action
		host = apiPod.Spec.NodeName
		status = m[action]
	} else if apiRC, ok := obj.(*v1.ReplicationController); ok {
		namespace = apiRC.ObjectMeta.Namespace
		name = apiRC.Name
		kind = "replication controller"
		reason = action
		status = m[action]
	} else if apiDeployment, ok := obj.(*v1beta1.Deployment); ok {
		namespace = apiDeployment.ObjectMeta.Namespace
		name = apiDeployment.Name
		kind = "deployment"
		reason = action
		status = m[action]
	} else if apiJob, ok := obj.(*batchv1.Job); ok {
		namespace = apiJob.ObjectMeta.Namespace
		name = apiJob.Name
		kind = "job"
		reason = action
		status = m[action]
	} else if apiPV, ok := obj.(*v1.PersistentVolume); ok {
		name = apiPV.Name
		kind = "persistent volume"
		reason = action
		status = m[action]
	} else if apiEvent, ok := obj.(*v1.Event); ok {
		namespace = apiEvent.Namespace
		name = fmt.Sprintf("%s @ %s", apiEvent.InvolvedObject.Name, apiEvent.Source.Host)
		kind = apiEvent.InvolvedObject.Kind
		reason = apiEvent.Reason
		status = apiEvent.Type
		message = action + " - " + apiEvent.Message
	} else {
		log.Printf("Unknown object type: %s -- value --> %s", reflect.TypeOf(obj), obj)
	}

	kbEvent := Event{
		Namespace: namespace,
		Kind:      kind,
		Component: component,
		Host:      host,
		Reason:    reason,
		Status:    status,
		Name:      name,
		Message:   message,
	}

	return kbEvent
}
