/*
Copyright 2016 Skippbox, Ltd.

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

package config

import (
	"fmt"
	"log"

	"github.com/edevil/kubewatch/pkg/handlers"
	"github.com/spf13/viper"
)

// Resource contains resource configuration
type Resource struct {
	Deployment            bool
	ReplicationController bool
	ReplicaSet            bool
	DaemonSet             bool
	Service               bool
	Pod                   bool
	Job                   bool
	PersistentVolume      bool
	Event                 bool
}

// Config struct contains kubewatch configuration
type Config struct {
	Resource    Resource
	InCluster   bool
	Handlers    []handlers.Handler
	InitialList bool
}

// Load loads configuration from config file
func (c *Config) Load() error {
	c.Resource.DaemonSet = viper.GetBool("resource.daemonset")
	c.Resource.Deployment = viper.GetBool("resource.deployment")
	c.Resource.Job = viper.GetBool("resource.job")
	c.Resource.PersistentVolume = viper.GetBool("resource.persistentvolume")
	c.Resource.Pod = viper.GetBool("resource.pod")
	c.Resource.ReplicaSet = viper.GetBool("resource.replicaset")
	c.Resource.ReplicationController = viper.GetBool("resource.replicationcontroller")
	c.Resource.Service = viper.GetBool("resource.service")
	c.Resource.Event = viper.GetBool("resource.event")

	c.InCluster = viper.GetBool("inCluster")

	handlerMap := viper.GetStringMap("handler")
	for hName := range handlerMap {
		if hObject, ok := handlers.HandlerMap[hName]; ok {
			if err := hObject.Init(viper.Sub("handler." + hName)); err != nil {
				return err
			}

			c.Handlers = append(c.Handlers, hObject)
			log.Println("Configured handler:", hName)
		} else {
			return fmt.Errorf("could not find handler: %s", hName)
		}
	}

	c.InitialList = viper.GetBool("initialList")

	return nil
}
