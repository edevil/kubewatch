/*
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

package handlers

import (
	"log"

	"github.com/spf13/viper"
)

// Default handler implements Handler interface,
// print each event with JSON format
type Default struct {
}

// Init initializes handler configuration
// Do nothing for default handler
func (d *Default) Init(c *viper.Viper) error {
	log.Println("Init called")
	return nil
}

func (d *Default) ObjectCreated(obj interface{}) {
	log.Println("Object created:", obj)
}

func (d *Default) ObjectDeleted(obj interface{}) {
	log.Println("Object deleted:", obj)
}

func (d *Default) ObjectUpdated(oldObj, newObj interface{}, changes []string) {
	log.Println("Object updated. Old:", oldObj, "New object:", newObj)
	for _, change := range changes {
		log.Println(change)
	}
}

func init() {
	HandlerMap["default"] = &Default{}
}
