package message

import "time"

// Message represents a message in Robust
type Message struct {
	ID        string            `json:"id,omitempty"`      // Unique message ID
	Topic     string            `json:"topic"`             // Topic name
	Partition int32             `json:"partition"`         // Partition number
	Key       []byte            `json:"key,omitempty"`     // Message key (for partitioning)
	Value     []byte            `json:"value"`             // Message value (payload)
	Headers   map[string]string `json:"headers,omitempty"` // Message headers
	Timestamp int64             `json:"timestamp"`         // Message timestamp (Unix milliseconds)
	Offset    int64             `json:"offset"`            // Message offset in partition
}

// NewMessage creates a new message with the specified topic and value
func NewMessage(topic string, value []byte) *Message {
	return &Message{
		Topic:     topic,
		Value:     value,
		Headers:   make(map[string]string),
		Timestamp: time.Now().UnixMilli(),
		Partition: -1, // Will be assigned by producer/broker
		Offset:    -1, // Will be assigned by broker
	}
}

// NewMessageWithKey creates a new message with topic, key, and value
func NewMessageWithKey(topic string, key, value []byte) *Message {
	return &Message{
		Topic:     topic,
		Key:       key,
		Value:     value,
		Headers:   make(map[string]string),
		Timestamp: time.Now().UnixMilli(),
		Partition: -1,
		Offset:    -1,
	}
}

// Size returns the approximate size of the message in bytes
func (m *Message) Size() int {
	size := len(m.ID) + len(m.Topic) + len(m.Key) + len(m.Value)
	size += 8 + 4 + 8 // Timestamp + Partition + Offset
	for k, v := range m.Headers {
		size += len(k) + len(v)
	}
	return size
}

// Clone creates a deep copy of the message
func (m *Message) Clone() *Message {
	clone := &Message{
		ID:        m.ID,
		Topic:     m.Topic,
		Partition: m.Partition,
		Timestamp: m.Timestamp,
		Offset:    m.Offset,
	}

	if m.Key != nil {
		clone.Key = make([]byte, len(m.Key))
		copy(clone.Key, m.Key)
	}

	if m.Value != nil {
		clone.Value = make([]byte, len(m.Value))
		copy(clone.Value, m.Value)
	}

	if m.Headers != nil {
		clone.Headers = make(map[string]string, len(m.Headers))
		for k, v := range m.Headers {
			clone.Headers[k] = v
		}
	}

	return clone
}
