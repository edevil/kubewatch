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
	"github.com/spf13/viper"
)

// Handler - hardcoded configured handler
type Handler struct {
	Slack Slack `json:"slack"`
}

// Resource contains resource configuration
type Resource struct {
	Deployment            bool `json:"deployment"`
	ReplicationController bool `json:"rc"`
	ReplicaSet            bool `json:"rs"`
	DaemonSet             bool `json:"ds"`
	Services              bool `json:"svc"`
	Pod                   bool `json:"po"`
	Job                   bool `json:"job"`
	PersistentVolume      bool `json:"pv"`
}

// Config struct contains kubewatch configuration
type Config struct {
	Handler   Handler  `json:"handler"`
	Resource  Resource `json:"resource"`
	InCluster bool     `json:"inCluster"`
}

// Slack contains slack configuration
type Slack struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
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
	c.Resource.Services = viper.GetBool("resource.services")

	c.InCluster = viper.GetBool("inCluster")

	c.Handler.Slack.Channel = viper.GetString("handler.slack.channel")
	c.Handler.Slack.Token = viper.GetString("handler.slack.token")

	return nil
}
