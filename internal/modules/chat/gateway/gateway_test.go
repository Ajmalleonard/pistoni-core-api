package gateway

import (
	"errors"
	"testing"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"pistoni/core/internal/modules/chat/dto"
	"pistoni/core/internal/modules/chat/queue"
)

type fakeSvc struct {
	getRoomErr        error
	createErr         error
	createTargetedErr error
	createReportErr   error
	createdBy         string
	targetedBy        string
	reportBy          string
}

func (f *fakeSvc) GetChatRoom(userID, id string) (any, error) {
	if f.getRoomErr != nil {
		return nil, f.getRoomErr
	}
	return map[string]string{"id": id, "userId": userID}, nil
}
func (f *fakeSvc) CreateMessage(userID string, in dto.CreateMessageDTO) (any, error) {
	f.createdBy = userID
	if f.createErr != nil {
		return nil, f.createErr
	}
	return map[string]any{"chatRoomId": in.ChatRoomID, "content": in.Content}, nil
}
func (f *fakeSvc) CreateTargetedMessage(userID string, in dto.CreateTargetedMessageDTO) (any, error) {
	f.targetedBy = userID
	if f.createTargetedErr != nil {
		return nil, f.createTargetedErr
	}
	return map[string]any{"chatRoomId": in.ChatRoomID, "content": in.Content, "targetType": in.TargetType}, nil
}
func (f *fakeSvc) CreateTargetedReport(userID string, in dto.CreateTargetedReportDTO) (any, error) {
	f.reportBy = userID
	if f.createReportErr != nil {
		return nil, f.createReportErr
	}
	return map[string]any{"chatRoomId": in.ChatRoomID, "content": in.Content, "reportCategory": in.ReportCategory}, nil
}

type emitted struct {
	event string
	data  []any
}

type fakePublisher struct {
	events []queue.SignalEvent
	err    error
}

func (p *fakePublisher) PublishSignal(event queue.SignalEvent) error {
	p.events = append(p.events, event)
	return p.err
}

type fakeClient struct {
	id      string
	joined  []string
	left    []string
	emitted []emitted
}

func (c *fakeClient) ID() string { return c.id }
func (c *fakeClient) Join(room string) {
	c.joined = append(c.joined, room)
}
func (c *fakeClient) Leave(room string) {
	c.left = append(c.left, room)
}
func (c *fakeClient) Emit(eventName string, v ...any) {
	c.emitted = append(c.emitted, emitted{event: eventName, data: v})
}

func TestJoinRoomChecksMembershipAndEmitsAck(t *testing.T) {
	svc := &fakeSvc{}
	g := &Gateway{sio: socketio.NewServer(nil), svc: svc}
	g.setUser("c1", "u1")
	c := &fakeClient{id: "c1"}

	g.handleJoinRoom(c, map[string]any{"chatRoomId": "room-1"})

	if len(c.joined) != 1 || c.joined[0] != "room-1" {
		t.Fatalf("expected join room-1, got %#v", c.joined)
	}
	if len(c.emitted) == 0 || c.emitted[0].event != "chat.ack" {
		t.Fatalf("expected first emit chat.ack got %#v", c.emitted)
	}
}

func TestSendUsesAuthenticatedUserAndBroadcastsAck(t *testing.T) {
	svc := &fakeSvc{}
	g := &Gateway{sio: socketio.NewServer(nil), svc: svc}
	g.setUser("c2", "user-2")
	c := &fakeClient{id: "c2"}

	g.handleSend(c, dto.CreateMessageDTO{ChatRoomID: "room-2", Content: "hello"})

	if svc.createdBy != "user-2" {
		t.Fatalf("expected create by user-2 got %s", svc.createdBy)
	}
	foundAck := false
	for _, e := range c.emitted {
		if e.event == "chat.ack" {
			foundAck = true
			break
		}
	}
	if !foundAck {
		t.Fatalf("expected chat.ack emitted, got %#v", c.emitted)
	}
}

func TestSendTargetedErrorEmitsChatError(t *testing.T) {
	svc := &fakeSvc{createTargetedErr: errors.New("forbidden")}
	g := &Gateway{sio: socketio.NewServer(nil), svc: svc}
	g.setUser("c3", "user-3")
	c := &fakeClient{id: "c3"}

	g.handleSendTargeted(c, dto.CreateTargetedMessageDTO{ChatRoomID: "room-3", TripID: "trip-1", Content: "dispatch", TargetType: "driver"})

	if len(c.emitted) == 0 || c.emitted[0].event != "chat.error" {
		t.Fatalf("expected chat.error got %#v", c.emitted)
	}
}

func TestSendReportCallsServiceWithAuthenticatedUser(t *testing.T) {
	svc := &fakeSvc{}
	g := &Gateway{sio: socketio.NewServer(nil), svc: svc}
	g.setUser("c4", "user-4")
	c := &fakeClient{id: "c4"}

	g.handleSendReport(c, dto.CreateTargetedReportDTO{ChatRoomID: "room-4", TripID: "trip-1", Content: "incident", TargetType: "bodyguard", ReportCategory: "SECURITY"})

	if svc.reportBy != "user-4" {
		t.Fatalf("expected report by user-4 got %s", svc.reportBy)
	}
}

func TestBuildSignalEnvelope(t *testing.T) {
	env := buildSignalEnvelope("chat.send_targeted", "u1", "trip-1", "room-1", "driver", []string{"d1", "d2"}, "hello")
	if env.EventType != "chat.send_targeted" || env.Sender != "u1" || env.TripID != "trip-1" || env.ChatRoomID != "room-1" {
		t.Fatalf("unexpected envelope: %#v", env)
	}
	if env.TargetType != "driver" || len(env.TargetIDs) != 2 || env.Content != "hello" {
		t.Fatalf("unexpected target/content in envelope: %#v", env)
	}
	if time.Since(env.Timestamp) > 2*time.Second {
		t.Fatalf("expected recent timestamp, got %v", env.Timestamp)
	}
}

func TestSendTargetedQueueFailureDoesNotBlockAck(t *testing.T) {
	svc := &fakeSvc{}
	pub := &fakePublisher{err: errors.New("redis down")}
	g := &Gateway{sio: socketio.NewServer(nil), svc: svc, publisher: pub}
	g.setUser("c5", "user-5")
	c := &fakeClient{id: "c5"}

	g.handleSendTargeted(c, dto.CreateTargetedMessageDTO{ChatRoomID: "room-5", TripID: "trip-5", Content: "dispatch", TargetType: "driver", TargetIDs: []string{"d1"}})

	if len(pub.events) != 1 {
		t.Fatalf("expected event published even if it fails, got %d", len(pub.events))
	}
	foundAck := false
	for _, e := range c.emitted {
		if e.event == "chat.ack" {
			foundAck = true
			break
		}
	}
	if !foundAck {
		t.Fatalf("expected ack despite queue failure, got %#v", c.emitted)
	}
}
