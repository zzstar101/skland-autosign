package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Message represents a collected log entry.
type Message struct {
	Text   string
	IsError bool
}

// Notifier collects messages and pushes them to multiple endpoints.
type Notifier interface {
	Collect(msg Message)
	Push(ctx context.Context) error
}

// WebhookNotifier posts collected messages as a single payload to multiple URLs.
type WebhookNotifier struct {
	urls     []string
	messages []Message
	client   *http.Client
}

// NewWebhookNotifier creates a new notifier for the given URLs.
func NewWebhookNotifier(urls []string) *WebhookNotifier {
	return &WebhookNotifier{
		urls: urls,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Collect appends a message to the buffer.
func (n *WebhookNotifier) Collect(msg Message) {
	n.messages = append(n.messages, msg)
}

// Push sends all collected messages to each configured URL.
func (n *WebhookNotifier) Push(ctx context.Context) error {
	if len(n.urls) == 0 || len(n.messages) == 0 {
		return nil
	}

	type payloadMessage struct {
		Text    string `json:"text"`
		IsError bool   `json:"isError"`
	}
	var payload struct {
		Title    string           `json:"title"`
		Messages []payloadMessage `json:"messages"`
	}
	payload.Title = "森空岛每日签到"
	for _, m := range n.messages {
		payload.Messages = append(payload.Messages, payloadMessage{
			Text:    m.Text,
			IsError: m.IsError,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, url := range n.urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		if _, err := n.client.Do(req); err != nil {
			return err
		}
	}

	return nil
}

