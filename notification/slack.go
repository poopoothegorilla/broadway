package notification

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
	"github.com/namely/broadway/env"
)

// Message represents a JSON payload sent to slack
// see https://api.slack.com/docs/attachments
type Message struct {
	ResponseType string       `json:"response_type"`
	Attachments  []Attachment `json:"attachments"`
	Text         string       `json:"text"`
}

// NewMessage crafts a new Slack message. If ephemeral is true, the message gets
// delivered only to the Slack user who requested it, otherwise it goes to a
// channel
func NewMessage(ephemeral bool, msg string) *Message {
	var rt string
	if ephemeral {
		rt = "ephemeral"
	} else {
		rt = "in_channel"
	}
	return &Message{
		ResponseType: rt,
		Text:         msg,
	}
}

// Field represents a single field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Attachment represents a single item in a Slack attachment payload array
type Attachment struct {
	Fallback   string   `json:"fallback"`
	Color      string   `json:"color"`
	Pretext    string   `json:"pretext"`
	AuthorName string   `json:"author_name"`
	AuthorLink string   `json:"author_link"`
	AuthorIcon string   `json:"author_icon"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link"`
	Text       string   `json:"text"`
	Fields     []Field  `json:"fields"`
	ImageURL   string   `json:"image_url"`
	ThumbURL   string   `json:"thumb_url"`
	MarkdownIn []string `json:"mrkdwn_in"` // valid options: "pretext", "text", "fields"
}

// Send sends the Slack notification
func (message *Message) Send() error {
	if env.SlackWebhook == "" {
		glog.Warningf("SLACK_WEBHOOK env var is unset, not sending %s", message.Text)
		return nil
	}

	value, err := json.Marshal(message)
	if err != nil {
		return err
	}
	glog.Infof("Sending Slack message to %s", env.SlackWebhook)
	resp, err := http.Post(env.SlackWebhook, "application/json", bytes.NewReader(value))
	glog.Info("Slack returned: ", resp)
	return err
}
