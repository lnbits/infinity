package nostr_utils

import (
	"fmt"
	"time"

	nostr "github.com/fiatjaf/go-nostr"
	"github.com/lnbits/infinity/events"
)

type EventRelay struct {
	Event *nostr.Event `json:"event"`
	Relay string       `json:"relay"`
}

func Publish(app, wallet string, eventData map[string]interface{}) (string, error) {
	evt := &nostr.Event{
		CreatedAt: time.Now(),
		Kind:      1,
		Tags:      make(nostr.Tags, 0),
		Content:   "",
	}

	if ikind, ok := eventData["kind"]; ok {
		if kind, ok := ikind.(int); ok {
			evt.Kind = kind
		} else {
			return "", fmt.Errorf("invalid kind: %v", ikind)
		}
	}

	if itags, ok := eventData["tags"]; ok {
		if tags, ok := itags.([]interface{}); ok {
			for i, itag := range tags {
				if tag, ok := itag.(nostr.StringList); ok {
					evt.Tags = append(evt.Tags, tag)
				} else {
					return "", fmt.Errorf("invalid tag at %d: %v", i, itag)
				}
			}
		} else {
			return "", fmt.Errorf("invalid tags: %v", itags)
		}
	}

	if icontent, ok := eventData["content"]; ok {
		if content, ok := icontent.(string); ok {
			evt.Content = content
		} else {
			return "", fmt.Errorf("invalid content: %v", icontent)
		}
	}

	evt, statuses, err := pool.PublishEvent(evt)
	if err != nil {
		return "", err
	}

	go func() {
		for status := range statuses {
			switch status.Status {
			case nostr.PublishStatusSent:
				events.EmitGenericAppWalletEvent(
					app, wallet,
					"nostr_event_sent",
					EventRelay{evt, status.Relay},
				)
			case nostr.PublishStatusFailed:
				events.EmitGenericAppWalletEvent(
					app, wallet,
					"nostr_event_failed",
					EventRelay{evt, status.Relay},
				)
			case nostr.PublishStatusSucceeded:
				events.EmitGenericAppWalletEvent(
					app, wallet,
					"nostr_event_confirmed",
					EventRelay{evt, status.Relay},
				)
			}
		}
	}()

	return evt.ID, nil
}
