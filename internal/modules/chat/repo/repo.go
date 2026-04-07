package repo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/modules/chat/dto"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Repository interface {
	GetUserChatRooms(userID string) ([]model.ChatRoom, error)
	GetChatRoom(userID, chatRoomID string) (model.ChatRoom, error)
	GetMessages(chatRoomID string) ([]model.Message, error)
	CreateMessage(userID string, in dto.CreateMessageDTO) (model.Message, error)
	CreateTargetedMessage(userID string, in dto.CreateTargetedMessageDTO) (model.Message, error)
	CreateTargetedReport(userID string, in dto.CreateTargetedReportDTO) (model.Message, error)
	GetCommandSuggestions(tripID, q string) ([]dto.CommandSuggestion, error)
	GetCommandTargets(userID, tripID, targetType string) ([]dto.CommandTarget, error)
	EditMessage(userID string, in dto.EditMessageDTO) (model.Message, error)
	DeleteMessage(userID string, in dto.DeleteMessageDTO) (any, error)
	CreatePrivateChat(userID, participantID string) (model.ChatRoom, error)
	CreateGroupChat(creatorID, name string, participantIDs []string) (model.ChatRoom, error)
	GetUserByEmail(email string) (model.PublicUser, error)
}

type MemoryRepository struct {
	store *storage.MemoryStore
	users usersrepo.Repository
}

func NewMemoryRepository(store *storage.MemoryStore, users usersrepo.Repository) *MemoryRepository {
	return &MemoryRepository{store: store, users: users}
}

func (r *MemoryRepository) GetUserChatRooms(userID string) ([]model.ChatRoom, error) {
	r.store.Mu.RLock()
	out := []model.ChatRoom{}
	for _, room := range r.store.ChatRooms {
		if contains(room.ParticipantIDs, userID) {
			out = append(out, room)
		}
	}
	r.store.Mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (r *MemoryRepository) GetChatRoom(userID, chatRoomID string) (model.ChatRoom, error) {
	r.store.Mu.RLock()
	room, ok := r.store.ChatRooms[chatRoomID]
	r.store.Mu.RUnlock()
	if !ok || !contains(room.ParticipantIDs, userID) {
		return model.ChatRoom{}, apperr.New(404, "Chat room not found or you are not a participant")
	}
	return room, nil
}

func (r *MemoryRepository) GetMessages(chatRoomID string) ([]model.Message, error) {
	r.store.Mu.RLock()
	ids := append([]string{}, r.store.MessagesByRoom[chatRoomID]...)
	out := make([]model.Message, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.store.Messages[id])
	}
	r.store.Mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp.Before(out[j].Timestamp) })
	return out, nil
}

func (r *MemoryRepository) CreateMessage(userID string, in dto.CreateMessageDTO) (model.Message, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	room, ok := r.store.ChatRooms[in.ChatRoomID]
	if !ok {
		return model.Message{}, apperr.New(404, "Chat room not found")
	}
	if !contains(room.ParticipantIDs, userID) {
		return model.Message{}, apperr.New(403, "You are not a participant of this chat room")
	}
	now := time.Now()
	m := model.Message{ID: uuid.NewString(), ChatRoomID: in.ChatRoomID, SenderID: userID, Content: in.Content, Timestamp: now}
	r.store.Messages[m.ID] = m
	r.store.MessagesByRoom[in.ChatRoomID] = append(r.store.MessagesByRoom[in.ChatRoomID], m.ID)
	room.LastMessageID = &m.ID
	room.UpdatedAt = now
	r.store.ChatRooms[room.ID] = room
	return m, nil
}

func (r *MemoryRepository) CreateTargetedMessage(userID string, in dto.CreateTargetedMessageDTO) (model.Message, error) {
	target, err := r.resolveTarget(userID, in.TripID, in.TargetType, in.TargetIDs, in.BodyguardIndex)
	if err != nil {
		return model.Message{}, err
	}
	if err := r.validateTargetedOpsPermission(userID, in.ChatRoomID); err != nil {
		return model.Message{}, err
	}
	return r.createMessageWithTarget(userID, in.ChatRoomID, in.Content, target, nil)
}

func (r *MemoryRepository) CreateTargetedReport(userID string, in dto.CreateTargetedReportDTO) (model.Message, error) {
	r.store.Mu.RLock()
	rental, ok := r.store.Rentals[in.TripID]
	r.store.Mu.RUnlock()
	if !ok {
		return model.Message{}, apperr.New(404, "Trip not found")
	}
	if rental.UserID != userID {
		return model.Message{}, apperr.New(403, "Only trip client can send targeted reports")
	}
	target, err := r.resolveTarget(userID, in.TripID, in.TargetType, in.TargetIDs, in.BodyguardIndex)
	if err != nil {
		return model.Message{}, err
	}
	return r.createMessageWithTarget(userID, in.ChatRoomID, in.Content, target, &in.ReportCategory)
}

func (r *MemoryRepository) GetCommandSuggestions(tripID, q string) ([]dto.CommandSuggestion, error) {
	_ = tripID
	options := []dto.CommandSuggestion{
		{Command: "/", Label: "Open slash commands"},
		{Command: "/driver", Label: "Target assigned driver"},
		{Command: "/bodyguard", Label: "Target one bodyguard or a group"},
		{Command: "/report", Label: "Send incident/security report"},
	}
	query := strings.ToLower(strings.TrimSpace(q))
	if query == "" || query == "/" {
		return options, nil
	}
	filtered := make([]dto.CommandSuggestion, 0)
	for _, op := range options {
		if strings.Contains(strings.ToLower(op.Command), query) || strings.Contains(strings.ToLower(op.Label), query) {
			filtered = append(filtered, op)
		}
	}
	return filtered, nil
}

func (r *MemoryRepository) GetCommandTargets(userID, tripID, targetType string) ([]dto.CommandTarget, error) {
	r.store.Mu.RLock()
	rental, ok := r.store.Rentals[tripID]
	r.store.Mu.RUnlock()
	if !ok {
		return nil, apperr.New(404, "Trip not found")
	}
	if !isTripActor(rental, userID) {
		return nil, apperr.New(403, "Forbidden trip access")
	}
	switch targetType {
	case "driver":
		if rental.DriverID == nil || *rental.DriverID == "" {
			return []dto.CommandTarget{}, nil
		}
		u, _ := r.users.FindByID(*rental.DriverID)
		label := "Driver"
		if u.ID != "" {
			label = fmt.Sprintf("Driver · %s", u.Name)
		}
		return []dto.CommandTarget{{ID: *rental.DriverID, Label: label, Type: "driver"}}, nil
	case "bodyguard":
		out := make([]dto.CommandTarget, 0, len(rental.BodyGuardIDs))
		for i, id := range rental.BodyGuardIDs {
			u, _ := r.users.FindByID(id)
			label := fmt.Sprintf("Bodyguard #%d", i+1)
			if u.ID != "" {
				label = fmt.Sprintf("Bodyguard #%d · %s", i+1, u.Name)
			}
			out = append(out, dto.CommandTarget{ID: id, Label: label, Type: "bodyguard", Index: i + 1})
		}
		return out, nil
	case "bodyguard-group":
		if len(rental.BodyGuardIDs) == 0 {
			return []dto.CommandTarget{}, nil
		}
		return []dto.CommandTarget{{ID: "all", Label: "All bodyguards", Type: "bodyguard-group"}}, nil
	default:
		return nil, apperr.New(400, "Unsupported target type")
	}
}

func (r *MemoryRepository) EditMessage(userID string, in dto.EditMessageDTO) (model.Message, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	m, ok := r.store.Messages[in.MessageID]
	if !ok || m.ChatRoomID != in.ChatRoomID {
		return model.Message{}, apperr.New(404, "Message not found")
	}
	if m.SenderID != userID {
		return model.Message{}, apperr.New(403, "You can only edit your own messages")
	}
	m.Content = in.Content
	m.Edited = true
	r.store.Messages[in.MessageID] = m
	return m, nil
}

func (r *MemoryRepository) DeleteMessage(userID string, in dto.DeleteMessageDTO) (any, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	m, ok := r.store.Messages[in.MessageID]
	if !ok || m.ChatRoomID != in.ChatRoomID {
		return nil, apperr.New(404, "Message not found")
	}
	room := r.store.ChatRooms[in.ChatRoomID]
	isAdmin := room.Type == "GROUP" && len(room.ParticipantIDs) > 0 && room.ParticipantIDs[0] == userID
	if m.SenderID != userID && !isAdmin {
		return nil, apperr.New(403, "You cannot delete this message")
	}
	if in.HardDelete && isAdmin {
		delete(r.store.Messages, in.MessageID)
		ids := r.store.MessagesByRoom[in.ChatRoomID]
		out := ids[:0]
		for _, id := range ids {
			if id != in.MessageID {
				out = append(out, id)
			}
		}
		r.store.MessagesByRoom[in.ChatRoomID] = out
		return map[string]any{"id": in.MessageID, "deleted": true}, nil
	}
	m.IsDeleted = true
	m.Content = "This message was deleted"
	r.store.Messages[in.MessageID] = m
	return m, nil
}

func (r *MemoryRepository) CreatePrivateChat(userID, participantID string) (model.ChatRoom, error) {
	if _, err := r.users.FindByID(participantID); err != nil {
		return model.ChatRoom{}, apperr.New(404, "User not found")
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	for _, room := range r.store.ChatRooms {
		if room.Type == "PRIVATE" && contains(room.ParticipantIDs, userID) && contains(room.ParticipantIDs, participantID) {
			return room, nil
		}
	}
	now := time.Now()
	room := model.ChatRoom{ID: uuid.NewString(), Type: "PRIVATE", ParticipantIDs: []string{userID, participantID}, CreatedAt: now, UpdatedAt: now}
	r.store.ChatRooms[room.ID] = room
	return room, nil
}

func (r *MemoryRepository) CreateGroupChat(creatorID, name string, participantIDs []string) (model.ChatRoom, error) {
	if !contains(participantIDs, creatorID) {
		participantIDs = append(participantIDs, creatorID)
	}
	for _, id := range participantIDs {
		if _, err := r.users.FindByID(id); err != nil {
			return model.ChatRoom{}, apperr.New(404, "One or more users not found")
		}
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	now := time.Now()
	n := name
	room := model.ChatRoom{ID: uuid.NewString(), Name: &n, Type: "GROUP", ParticipantIDs: participantIDs, CreatedAt: now, UpdatedAt: now}
	r.store.ChatRooms[room.ID] = room
	return room, nil
}

func (r *MemoryRepository) GetUserByEmail(email string) (model.PublicUser, error) {
	u, err := r.users.FindByEmail(email)
	if err != nil {
		return model.PublicUser{}, err
	}
	return model.PublicUser{ID: u.ID, Name: u.Name, PhoneNumber: u.PhoneNumber, ProfilePicture: u.ProfilePicture}, nil
}

func (r *MemoryRepository) createMessageWithTarget(userID, chatRoomID, content string, target dto.CommandTarget, reportCategory *string) (model.Message, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	room, ok := r.store.ChatRooms[chatRoomID]
	if !ok {
		return model.Message{}, apperr.New(404, "Chat room not found")
	}
	if !contains(room.ParticipantIDs, userID) {
		return model.Message{}, apperr.New(403, "You are not a participant of this chat room")
	}
	now := time.Now()
	typeVal := target.Type
	labelVal := target.Label
	m := model.Message{
		ID:             uuid.NewString(),
		ChatRoomID:     chatRoomID,
		SenderID:       userID,
		Content:        content,
		Timestamp:      now,
		TargetType:     &typeVal,
		TargetIDs:      []string{target.ID},
		TargetLabel:    &labelVal,
		ReportCategory: reportCategory,
	}
	if target.Type == "bodyguard-group" {
		m.TargetIDs = append([]string{}, strings.Split(target.ID, ",")...)
	}
	r.store.Messages[m.ID] = m
	r.store.MessagesByRoom[chatRoomID] = append(r.store.MessagesByRoom[chatRoomID], m.ID)
	room.LastMessageID = &m.ID
	room.UpdatedAt = now
	r.store.ChatRooms[room.ID] = room
	return m, nil
}

func (r *MemoryRepository) resolveTarget(userID, tripID, targetType string, targetIDs []string, bodyguardIndex *int) (dto.CommandTarget, error) {
	r.store.Mu.RLock()
	rental, ok := r.store.Rentals[tripID]
	r.store.Mu.RUnlock()
	if !ok {
		return dto.CommandTarget{}, apperr.New(404, "Trip not found")
	}
	if !isTripActor(rental, userID) {
		return dto.CommandTarget{}, apperr.New(403, "Forbidden trip access")
	}
	switch targetType {
	case "driver":
		if rental.DriverID == nil || *rental.DriverID == "" {
			return dto.CommandTarget{}, apperr.New(404, "Driver is not assigned")
		}
		return dto.CommandTarget{ID: *rental.DriverID, Type: "driver", Label: "Driver"}, nil
	case "bodyguard":
		if bodyguardIndex != nil {
			idx := *bodyguardIndex - 1
			if idx < 0 || idx >= len(rental.BodyGuardIDs) {
				return dto.CommandTarget{}, apperr.New(400, "Invalid bodyguard index")
			}
			id := rental.BodyGuardIDs[idx]
			return dto.CommandTarget{ID: id, Type: "bodyguard", Label: fmt.Sprintf("Bodyguard #%d", *bodyguardIndex), Index: *bodyguardIndex}, nil
		}
		if len(targetIDs) == 0 {
			return dto.CommandTarget{}, apperr.New(400, "bodyguard target requires bodyguardIndex or targetIds")
		}
		if !contains(rental.BodyGuardIDs, targetIDs[0]) {
			return dto.CommandTarget{}, apperr.New(400, "Target bodyguard not in trip")
		}
		return dto.CommandTarget{ID: targetIDs[0], Type: "bodyguard", Label: "Bodyguard"}, nil
	case "bodyguard-group":
		ids := []string{}
		if len(targetIDs) == 0 {
			ids = append(ids, rental.BodyGuardIDs...)
		} else {
			for _, id := range targetIDs {
				if !contains(rental.BodyGuardIDs, id) {
					return dto.CommandTarget{}, apperr.New(400, "One or more bodyguards not in trip")
				}
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return dto.CommandTarget{}, apperr.New(400, "No bodyguards available for group target")
		}
		return dto.CommandTarget{ID: strings.Join(ids, ","), Type: "bodyguard-group", Label: "Bodyguard group"}, nil
	default:
		return dto.CommandTarget{}, apperr.New(400, "Unsupported target type")
	}
}

func (r *MemoryRepository) validateTargetedOpsPermission(userID, chatRoomID string) error {
	u, err := r.users.FindByID(userID)
	if err != nil {
		return apperr.New(401, "Unauthorized")
	}
	if u.Role != model.UserRoleDriver && u.Role != model.UserRoleBodyGuard {
		return nil
	}
	r.store.Mu.RLock()
	room, ok := r.store.ChatRooms[chatRoomID]
	r.store.Mu.RUnlock()
	if !ok {
		return apperr.New(404, "Chat room not found")
	}
	if room.Type != "GROUP" {
		return apperr.New(403, "Driver/bodyguard targeted messages are only allowed in ops rooms")
	}
	return nil
}

func isTripActor(r model.Rental, userID string) bool {
	if r.UserID == userID {
		return true
	}
	if r.DriverID != nil && *r.DriverID == userID {
		return true
	}
	for _, id := range r.BodyGuardIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func contains(v []string, id string) bool {
	for _, x := range v {
		if x == id {
			return true
		}
	}
	return false
}
