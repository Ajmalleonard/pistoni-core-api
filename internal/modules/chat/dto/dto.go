package dto

type CreateMessageDTO struct {
	ChatRoomID string `json:"chatRoomId" binding:"required"`
	Content    string `json:"content" binding:"required"`
}
type EditMessageDTO struct {
	ChatRoomID string `json:"chatRoomId" binding:"required"`
	MessageID  string `json:"messageId"`
	Content    string `json:"content" binding:"required"`
}
type DeleteMessageDTO struct {
	ChatRoomID string `json:"chatRoomId" binding:"required"`
	MessageID  string `json:"messageId"`
	HardDelete bool   `json:"hardDelete"`
}
type CreateChatRoomDTO struct {
	ParticipantID  string   `json:"participantId"`
	Name           string   `json:"name"`
	ParticipantIDs []string `json:"participantIds"`
	Type           string   `json:"type" binding:"required"`
}

type CommandSuggestion struct {
	Command string `json:"command"`
	Label   string `json:"label"`
}

type CommandTarget struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
}

type CreateTargetedMessageDTO struct {
	TripID         string   `json:"tripId" binding:"required"`
	ChatRoomID     string   `json:"chatRoomId" binding:"required"`
	Content        string   `json:"content" binding:"required"`
	TargetType     string   `json:"targetType" binding:"required"`
	TargetIDs      []string `json:"targetIds"`
	BodyguardIndex *int     `json:"bodyguardIndex"`
}

type CreateTargetedReportDTO struct {
	TripID         string   `json:"tripId" binding:"required"`
	ChatRoomID     string   `json:"chatRoomId" binding:"required"`
	Content        string   `json:"content" binding:"required"`
	TargetType     string   `json:"targetType" binding:"required"`
	TargetIDs      []string `json:"targetIds"`
	BodyguardIndex *int     `json:"bodyguardIndex"`
	ReportCategory string   `json:"reportCategory" binding:"required"`
}
