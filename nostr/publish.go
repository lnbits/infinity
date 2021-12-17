package nostr

import (
	"fmt"
	"time"

	"github.com/fiatjaf/go-nostr/event"
)

func Publish(eventData map[string]interface{}) (string, error) {
	event := &event.Event{
		CreatedAt: uint32(time.Now().Unix()),
		Kind:      1,
		Tags:      make(event.Tags, 0),
		Content:   "",
	}

	if ikind, ok := eventData["kind"]; ok {
		if kind, ok := ikind.(int); ok {
			event.Kind = kind
		} else {
			return "", fmt.Errorf("invalid kind: %v", ikind)
		}
	}

	if itags, ok := eventData["tags"]; ok {
		if tags, ok := itags.([]interface{}); ok {
			for i, itag := range tags {
				if tag, ok := itag.(event.Tag); ok {
					event.Tags = append(event.Tags, tag)
				} else {
					return "", fmt.Errorf("invalid tag at %d: %v", i, itag)
				}
			}
		} else {
			return "", fmt.Errorf("invalid tags: %v", itags)
		}
	}

	if icontent, ok := eventData["content"]; ok {
		if content, ok := icontent.(int); ok {
			event.Content = content
		} else {
			return "", fmt.Errorf("invalid content: %v", icontent)
		}
	}

	event, statuses, err := pool.PublishEvent(event)
	if err != nil {
		return "", err
	}

	// TODO: trigger event to a single app
	// go func() {
	// 	for status := range statuses {
	// 		switch status.Status {
	// 		case relaypool.PublishStatusSent:
	// 			fmt.Printf("Sent event %s to '%s'.\n", event.ID, status.Relay)
	// 		case relaypool.PublishStatusFailed:
	// 			fmt.Printf("Failed to send event %s to '%s'.\n", event.ID, status.Relay)
	// 		case relaypool.PublishStatusSucceeded:
	// 			fmt.Printf("Event seen %s on '%s'.\n", event.ID, status.Relay)
	// 		}
	// 	}
	// }()

	return event.ID, nil
}
