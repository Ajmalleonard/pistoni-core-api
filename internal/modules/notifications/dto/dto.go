package dto

type CreateNotificationDTO struct {
	UserID          string `json:"userId" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Body            string `json:"body" binding:"required"`
	Type            string `json:"type" binding:"required"`
	RelatedEntityID string `json:"relatedEntityId"`
	ImageURL        string `json:"imageUrl"`
	ActionURL       string `json:"actionUrl"`
}

type MarkNotificationReadDTO struct {
	NotificationIDs []string `json:"notificationIds" binding:"required"`
}
