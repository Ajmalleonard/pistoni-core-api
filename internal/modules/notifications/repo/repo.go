package repo

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/modules/notifications/dto"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Repository interface {
	CreateNotification(in dto.CreateNotificationDTO) (model.Notification, error)
	GetUserNotifications(userID string, page, limit int) (map[string]any, error)
	GetUnreadCount(userID string) (map[string]int, error)
	MarkAsRead(ids []string, userID string) (map[string]bool, error)
	MarkAllAsRead(userID string) (map[string]bool, error)
	DeleteNotification(id, userID string) (map[string]bool, error)
}

type MemoryRepository struct{ store *storage.MemoryStore }

func NewMemoryRepository(store *storage.MemoryStore) *MemoryRepository {
	return &MemoryRepository{store: store}
}
func (r *MemoryRepository) CreateNotification(in dto.CreateNotificationDTO) (model.Notification, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	now := time.Now()
	var rel, img, act *string
	if in.RelatedEntityID != "" {
		rel = &in.RelatedEntityID
	}
	if in.ImageURL != "" {
		img = &in.ImageURL
	}
	if in.ActionURL != "" {
		act = &in.ActionURL
	}
	n := model.Notification{ID: uuid.NewString(), UserID: in.UserID, Title: in.Title, Body: in.Body, Type: model.NotificationType(in.Type), RelatedEntityID: rel, ImageURL: img, ActionURL: act, IsRead: false, CreatedAt: now, UpdatedAt: now}
	r.store.Notifications[n.ID] = n
	r.store.NotificationByUser[n.UserID] = append(r.store.NotificationByUser[n.UserID], n.ID)
	return n, nil
}
func (r *MemoryRepository) GetUserNotifications(userID string, page, limit int) (map[string]any, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	skip := (page - 1) * limit
	r.store.Mu.RLock()
	ids := append([]string{}, r.store.NotificationByUser[userID]...)
	out := make([]model.Notification, 0, len(ids))
	unread := 0
	for _, id := range ids {
		n := r.store.Notifications[id]
		out = append(out, n)
		if !n.IsRead {
			unread++
		}
	}
	r.store.Mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	total := len(out)
	if skip > len(out) {
		out = []model.Notification{}
	} else {
		out = out[skip:]
		if limit < len(out) {
			out = out[:limit]
		}
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return map[string]any{"notifications": out, "total": total, "unreadCount": unread, "page": page, "limit": limit, "totalPages": totalPages}, nil
}
func (r *MemoryRepository) GetUnreadCount(userID string) (map[string]int, error) {
	r.store.Mu.RLock()
	c := 0
	for _, id := range r.store.NotificationByUser[userID] {
		if !r.store.Notifications[id].IsRead {
			c++
		}
	}
	r.store.Mu.RUnlock()
	return map[string]int{"unreadCount": c}, nil
}
func (r *MemoryRepository) MarkAsRead(ids []string, userID string) (map[string]bool, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	for _, id := range ids {
		n, ok := r.store.Notifications[id]
		if !ok || n.UserID != userID {
			return nil, apperr.New(404, "Some notifications not found")
		}
	}
	for _, id := range ids {
		n := r.store.Notifications[id]
		n.IsRead = true
		n.UpdatedAt = time.Now()
		r.store.Notifications[id] = n
	}
	return map[string]bool{"success": true}, nil
}
func (r *MemoryRepository) MarkAllAsRead(userID string) (map[string]bool, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	for _, id := range r.store.NotificationByUser[userID] {
		n := r.store.Notifications[id]
		if !n.IsRead {
			n.IsRead = true
			n.UpdatedAt = time.Now()
			r.store.Notifications[id] = n
		}
	}
	return map[string]bool{"success": true}, nil
}
func (r *MemoryRepository) DeleteNotification(id, userID string) (map[string]bool, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	n, ok := r.store.Notifications[id]
	if !ok || n.UserID != userID {
		return nil, apperr.New(404, "Notification not found")
	}
	delete(r.store.Notifications, id)
	ids := r.store.NotificationByUser[userID]
	out := ids[:0]
	for _, x := range ids {
		if x != id {
			out = append(out, x)
		}
	}
	r.store.NotificationByUser[userID] = out
	return map[string]bool{"success": true}, nil
}
