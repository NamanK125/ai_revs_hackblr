package mocks

import (
	"fmt"
	"sync"
	"time"

	"github.com/mihup/engineerops/internal/stream"
)

// Message represents a Slack message.
type Message struct {
	Channel   string
	User      string
	Text      string
	Timestamp int64 // Unix timestamp
}

// SlackMock is an in-memory mock of Slack operations with concurrent safety.
type SlackMock struct {
	mu       sync.Mutex
	messages []Message
	bus      *stream.Bus
}

// validChannels is the set of allowed Slack channels.
var validChannels = map[string]bool{
	"#engineering": true,
	"#oncall":      true,
	"#random":      true,
}

// NewSlackMock creates a new Slack mock.
// bus may be nil for testing; mutations will publish events if bus is present.
func NewSlackMock(bus *stream.Bus) *SlackMock {
	return &SlackMock{
		messages: []Message{},
		bus:      bus,
	}
}

// Post posts a message to a channel.
// Returns error if the channel is unknown.
// Publishes to SSE bus if configured.
func (s *SlackMock) Post(channel, user, text string) (Message, error) {
	if !validChannels[channel] {
		return Message{}, fmt.Errorf("unknown channel: %%s", channel)
	}

	s.mu.Lock()
	msg := Message{
		Channel:   channel,
		User:      user,
		Text:      text,
		Timestamp: time.Now().Unix(),
	}
	s.messages = append(s.messages, msg)
	s.mu.Unlock()

	// Publish to SSE bus if available
	if s.bus != nil {
		 s.bus.Publish("slack_message_posted", map[string]interface{}{
			"channel":   msg.Channel,
			"user":      msg.User,
			"text":      msg.Text,
			"timestamp": msg.Timestamp,
		})
	}

	return msg, nil
}

// Recent returns the most recent messages in a channel, up to limit.
func (s *SlackMock) Recent(channel string, limit int) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []Message
	// Iterate backwards from the end to get most recent first
	for i := len(s.messages) - 1; i >= 0 && len(result) < limit; i-- {
		if s.messages[i].Channel == channel {
			result = append(result, s.messages[i])
		}
	}
	return result
}