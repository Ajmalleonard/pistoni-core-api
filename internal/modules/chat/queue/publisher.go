package queue

import "time"

type SignalEvent struct {
	EventType  string    `json:"eventType"`
	TripID     string    `json:"tripId"`
	ChatRoomID string    `json:"chatRoomId"`
	Sender     string    `json:"sender"`
	TargetType string    `json:"targetType"`
	TargetIDs  []string  `json:"targetIds"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

type QueuePublisher interface {
	PublishSignal(event SignalEvent) error
}

type NoopPublisher struct{}

func NewNoopPublisher() *NoopPublisher { return &NoopPublisher{} }

func (n *NoopPublisher) PublishSignal(event SignalEvent) error { return nil }
