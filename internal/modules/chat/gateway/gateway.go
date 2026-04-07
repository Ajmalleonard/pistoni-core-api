package gateway

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	socketio "github.com/googollee/go-socket.io"
	"pistoni/core/internal/modules/chat/dto"
	"pistoni/core/internal/modules/chat/queue"
	"pistoni/core/internal/modules/chat/service"
	"pistoni/core/internal/platform/auth"
)

type chatService interface {
	GetChatRoom(userID, id string) (any, error)
	CreateMessage(userID string, in dto.CreateMessageDTO) (any, error)
	CreateTargetedMessage(userID string, in dto.CreateTargetedMessageDTO) (any, error)
	CreateTargetedReport(userID string, in dto.CreateTargetedReportDTO) (any, error)
}

type socketClient interface {
	ID() string
	Join(room string)
	Leave(room string)
	Emit(eventName string, v ...any)
}

type connState struct {
	UserID string
}

type Gateway struct {
	sio       *socketio.Server
	svc       chatService
	jwt       *auth.JWTManager
	path      string
	states    sync.Map // connID -> connState
	publisher queue.QueuePublisher
}

func New(svc *service.Service, jwt *auth.JWTManager, path string, publisher queue.QueuePublisher) (*Gateway, error) {
	if path == "" {
		path = "/socket.io"
	}
	if publisher == nil {
		publisher = queue.NewNoopPublisher()
	}

	sio := socketio.NewServer(nil)
	g := &Gateway{sio: sio, svc: svc, jwt: jwt, path: path, publisher: publisher}
	g.registerHandlers()

	return g, nil
}

func (g *Gateway) Server() *socketio.Server { return g.sio }
func (g *Gateway) Path() string             { return g.path }

func (g *Gateway) registerHandlers() {
	g.sio.OnConnect("/", func(c socketio.Conn) error {
		uid, err := g.resolveUserID(c.URL())
		if err != nil {
			c.Emit("chat.error", map[string]any{"message": err.Error()})
			return err
		}
		g.setUser(c.ID(), uid)
		c.SetContext(connState{UserID: uid})
		c.Emit("chat.ack", map[string]any{"event": "connection/auth", "status": "ok", "userId": uid})
		c.Emit("presence.update", map[string]any{"userId": uid, "status": "online", "connectionId": c.ID()})
		return nil
	})

	g.sio.OnEvent("/", "connection/auth", func(c socketio.Conn, payload map[string]any) {
		uid, err := g.resolveUserFromPayload(payload)
		if err != nil {
			g.emitErr(c, "connection/auth", err)
			return
		}
		g.setUser(c.ID(), uid)
		c.SetContext(connState{UserID: uid})
		c.Emit("chat.ack", map[string]any{"event": "connection/auth", "status": "ok", "userId": uid})
		c.Emit("presence.update", map[string]any{"userId": uid, "status": "online", "connectionId": c.ID()})
	})

	g.sio.OnEvent("/", "chat.join_room", func(c socketio.Conn, payload map[string]any) {
		g.handleJoinRoom(realConn{Conn: c}, payload)
	})

	g.sio.OnEvent("/", "chat.leave_room", func(c socketio.Conn, payload map[string]any) {
		g.handleLeaveRoom(realConn{Conn: c}, payload)
	})

	g.sio.OnEvent("/", "chat.send", func(c socketio.Conn, payload dto.CreateMessageDTO) {
		g.handleSend(realConn{Conn: c}, payload)
	})

	g.sio.OnEvent("/", "chat.send_targeted", func(c socketio.Conn, payload dto.CreateTargetedMessageDTO) {
		g.handleSendTargeted(realConn{Conn: c}, payload)
	})

	g.sio.OnEvent("/", "chat.send_report", func(c socketio.Conn, payload dto.CreateTargetedReportDTO) {
		g.handleSendReport(realConn{Conn: c}, payload)
	})

	g.sio.OnDisconnect("/", func(c socketio.Conn, reason string) {
		uid := g.userIDFor(c)
		g.sio.BroadcastToNamespace("/", "presence.update", map[string]any{"userId": uid, "status": "offline", "connectionId": c.ID(), "reason": reason})
		g.states.Delete(c.ID())
	})

	g.sio.OnError("/", func(c socketio.Conn, err error) {
		log.Printf("[socketio] connection error conn=%s err=%v", c.ID(), err)
	})
}

type realConn struct{ socketio.Conn }

func (g *Gateway) handleJoinRoom(c socketClient, payload map[string]any) {
	uid, err := g.userIDForClient(c)
	if err != nil {
		g.emitErr(c, "chat.join_room", err)
		return
	}
	roomID, _ := payload["chatRoomId"].(string)
	if roomID == "" {
		g.emitErr(c, "chat.join_room", errors.New("chatRoomId is required"))
		return
	}
	if _, err := g.svc.GetChatRoom(uid, roomID); err != nil {
		g.emitErr(c, "chat.join_room", err)
		return
	}
	c.Join(roomID)
	c.Emit("chat.ack", map[string]any{"event": "chat.join_room", "status": "ok", "chatRoomId": roomID})
	g.sio.BroadcastToRoom("/", roomID, "presence.update", map[string]any{"chatRoomId": roomID, "userId": uid, "status": "joined"})
}

func (g *Gateway) handleLeaveRoom(c socketClient, payload map[string]any) {
	uid, err := g.userIDForClient(c)
	if err != nil {
		g.emitErr(c, "chat.leave_room", err)
		return
	}
	roomID, _ := payload["chatRoomId"].(string)
	if roomID == "" {
		g.emitErr(c, "chat.leave_room", errors.New("chatRoomId is required"))
		return
	}
	c.Leave(roomID)
	c.Emit("chat.ack", map[string]any{"event": "chat.leave_room", "status": "ok", "chatRoomId": roomID})
	g.sio.BroadcastToRoom("/", roomID, "presence.update", map[string]any{"chatRoomId": roomID, "userId": uid, "status": "left"})
}

func (g *Gateway) handleSend(c socketClient, payload dto.CreateMessageDTO) {
	uid, err := g.userIDForClient(c)
	if err != nil {
		g.emitErr(c, "chat.send", err)
		return
	}
	msg, err := g.svc.CreateMessage(uid, payload)
	if err != nil {
		g.emitErr(c, "chat.send", err)
		return
	}
	g.sio.BroadcastToRoom("/", payload.ChatRoomID, "chat.message", msg)
	c.Emit("chat.ack", map[string]any{"event": "chat.send", "status": "ok"})
}

func (g *Gateway) handleSendTargeted(c socketClient, payload dto.CreateTargetedMessageDTO) {
	uid, err := g.userIDForClient(c)
	if err != nil {
		g.emitErr(c, "chat.send_targeted", err)
		return
	}
	msg, err := g.svc.CreateTargetedMessage(uid, payload)
	if err != nil {
		g.emitErr(c, "chat.send_targeted", err)
		return
	}
	if err := g.publishSignal(buildSignalEnvelope("chat.send_targeted", uid, payload.TripID, payload.ChatRoomID, payload.TargetType, payload.TargetIDs, payload.Content)); err != nil {
		log.Printf("[chat.queue] warn: publish failed event=chat.send_targeted chatRoomId=%s err=%v", payload.ChatRoomID, err)
	}
	g.sio.BroadcastToRoom("/", payload.ChatRoomID, "chat.message", msg)
	c.Emit("chat.ack", map[string]any{"event": "chat.send_targeted", "status": "ok"})
}

func (g *Gateway) handleSendReport(c socketClient, payload dto.CreateTargetedReportDTO) {
	uid, err := g.userIDForClient(c)
	if err != nil {
		g.emitErr(c, "chat.send_report", err)
		return
	}
	msg, err := g.svc.CreateTargetedReport(uid, payload)
	if err != nil {
		g.emitErr(c, "chat.send_report", err)
		return
	}
	if err := g.publishSignal(buildSignalEnvelope("chat.send_report", uid, payload.TripID, payload.ChatRoomID, payload.TargetType, payload.TargetIDs, payload.Content)); err != nil {
		log.Printf("[chat.queue] warn: publish failed event=chat.send_report chatRoomId=%s err=%v", payload.ChatRoomID, err)
	}
	g.sio.BroadcastToRoom("/", payload.ChatRoomID, "chat.message", msg)
	c.Emit("chat.ack", map[string]any{"event": "chat.send_report", "status": "ok"})
}

func buildSignalEnvelope(eventType, sender, tripID, chatRoomID, targetType string, targetIDs []string, content string) queue.SignalEvent {
	return queue.SignalEvent{
		EventType:  eventType,
		TripID:     tripID,
		ChatRoomID: chatRoomID,
		Sender:     sender,
		TargetType: targetType,
		TargetIDs:  targetIDs,
		Content:    content,
		Timestamp:  time.Now().UTC(),
	}
}

func (g *Gateway) publishSignal(event queue.SignalEvent) error {
	if g.publisher == nil {
		return nil
	}
	return g.publisher.PublishSignal(event)
}

func (g *Gateway) emitErr(c socketClient, event string, err error) {
	c.Emit("chat.error", map[string]any{"event": event, "message": err.Error()})
}

func (g *Gateway) resolveUserID(u url.URL) (string, error) {
	q := u.Query()
	if uid := q.Get("userId"); uid != "" {
		return uid, nil
	}
	token := q.Get("token")
	if token == "" {
		return "", errors.New("missing userId/token in handshake")
	}
	if g.jwt == nil {
		return "", errors.New("token auth is not configured")
	}
	claims, err := g.jwt.Parse(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	if claims.Sub == "" {
		return "", errors.New("token missing subject")
	}
	return claims.Sub, nil
}

func (g *Gateway) resolveUserFromPayload(payload map[string]any) (string, error) {
	if payload == nil {
		return "", errors.New("auth payload is required")
	}
	if uid, _ := payload["userId"].(string); uid != "" {
		return uid, nil
	}
	token, _ := payload["token"].(string)
	if token == "" {
		return "", errors.New("missing userId/token in auth payload")
	}
	if g.jwt == nil {
		return "", errors.New("token auth is not configured")
	}
	claims, err := g.jwt.Parse(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	if claims.Sub == "" {
		return "", errors.New("token missing subject")
	}
	return claims.Sub, nil
}

func (g *Gateway) setUser(connID, userID string) {
	g.states.Store(connID, connState{UserID: userID})
}

func (g *Gateway) userIDFor(c socketio.Conn) string {
	if s, ok := c.Context().(connState); ok && s.UserID != "" {
		return s.UserID
	}
	if s, ok := g.states.Load(c.ID()); ok {
		if cs, ok := s.(connState); ok {
			return cs.UserID
		}
	}
	return ""
}

func (g *Gateway) userIDForClient(c socketClient) (string, error) {
	if s, ok := g.states.Load(c.ID()); ok {
		if cs, ok := s.(connState); ok && cs.UserID != "" {
			return cs.UserID, nil
		}
	}
	return "", errors.New("unauthenticated socket connection")
}
