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

package handlers

import (
	"fmt"
	"strings"

	"log"

	"github.com/spf13/viper"

	"github.com/nlopes/slack"

	"github.com/edevil/kubewatch/pkg/event"
)

var slackColors = map[string]string{
	"Normal":  "good",
	"Warning": "warning",
	"Danger":  "danger",
}

var slackErrMsg = `
%s

You need to set both slack token and channel for slack notify
in your configuration file.

`

// Slack handler implements handler.Handler interface,
// Notify event to slack channel
type Slack struct {
	Token   string
	Channel string
}

// Init prepares slack configuration
func (s *Slack) Init(c *viper.Viper) error {
	s.Token = c.GetString("token")
	s.Channel = c.GetString("channel")

	if s.Token == "" || s.Channel == "" {
		return fmt.Errorf(slackErrMsg, "Missing slack token or channel")
	}

	return nil
}

// ObjectCreated - implementation of Handler interface
func (s *Slack) ObjectCreated(obj interface{}) {
	notifySlack(s, obj, "created", "")
}

// ObjectDeleted - implementation of Handler interface
func (s *Slack) ObjectDeleted(obj interface{}) {
	notifySlack(s, obj, "deleted", "")
}

// ObjectUpdated - implementation of Handler interface
func (s *Slack) ObjectUpdated(oldObj, newObj interface{}, changes []string) {
	notifySlack(s, newObj, "updated", strings.Join(changes, ", "))
}

func notifySlack(s *Slack, obj interface{}, action string, extra string) {
	e := event.New(obj, action)
	api := slack.New(s.Token)
	params := slack.PostMessageParameters{}
	attachment := prepareSlackAttachment(e)

	params.Attachments = []slack.Attachment{attachment}
	params.AsUser = true
	channelID, timestamp, err := api.PostMessage(s.Channel, extra, params)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func prepareSlackAttachment(e event.Event) slack.Attachment {
	msg := fmt.Sprintf(
		"A %s in namespace %s has been %s: %s",
		e.Kind,
		e.Namespace,
		e.Reason,
		e.Name,
	)

	attachment := slack.Attachment{
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "kubewatch",
				Value: msg,
			},
		},
	}

	if color, ok := slackColors[e.Status]; ok {
		attachment.Color = color
	}

	attachment.MarkdownIn = []string{"fields"}

	return attachment
}

func init() {
	HandlerMap["slack"] = &Slack{}
}
