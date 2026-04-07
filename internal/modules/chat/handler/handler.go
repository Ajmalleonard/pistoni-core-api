package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/chat/dto"
	"pistoni/core/internal/modules/chat/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(s *service.Service) *Handler { return &Handler{svc: s} }
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/rooms", h.getChatRooms)
	r.GET("/rooms/:id", h.getChatRoom)
	r.GET("/rooms/:id/messages", h.getMessages)
	r.POST("/messages", h.createMessage)
	r.POST("/messages/targeted", h.createTargetedMessage)
	r.POST("/reports/targeted", h.createTargetedReport)
	r.GET("/commands/suggestions", h.getCommandSuggestions)
	r.GET("/commands/targets", h.getCommandTargets)
	r.PUT("/messages/:id", h.editMessage)
	r.DELETE("/messages/:id", h.deleteMessage)
	r.POST("/rooms", h.createChatRoom)
	r.GET("/users/email/:email", h.getUserByEmail)
}
func uid(c *gin.Context) string {
	v, _ := c.Get("userID")
	s, _ := v.(string)
	if s == "" {
		s = "test-user-123"
	}
	return s
}
func (h *Handler) getChatRooms(c *gin.Context) {
	d, e := h.svc.GetUserChatRooms(uid(c))
	respond(c, d, e)
}
func (h *Handler) getChatRoom(c *gin.Context) {
	d, e := h.svc.GetChatRoom(uid(c), c.Param("id"))
	respond(c, d, e)
}
func (h *Handler) getMessages(c *gin.Context) {
	d, e := h.svc.GetMessages(c.Param("id"))
	respond(c, d, e)
}
func (h *Handler) createMessage(c *gin.Context) {
	var in dto.CreateMessageDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateMessage(uid(c), in) })
}
func (h *Handler) createTargetedMessage(c *gin.Context) {
	var in dto.CreateTargetedMessageDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateTargetedMessage(uid(c), in) })
}
func (h *Handler) createTargetedReport(c *gin.Context) {
	var in dto.CreateTargetedReportDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateTargetedReport(uid(c), in) })
}
func (h *Handler) getCommandSuggestions(c *gin.Context) {
	d, e := h.svc.GetCommandSuggestions(c.Query("tripId"), c.Query("q"))
	respond(c, d, e)
}
func (h *Handler) getCommandTargets(c *gin.Context) {
	d, e := h.svc.GetCommandTargets(uid(c), c.Query("tripId"), c.Query("type"))
	respond(c, d, e)
}
func (h *Handler) editMessage(c *gin.Context) {
	var in dto.EditMessageDTO
	bind(c, &in, func() (any, error) { in.MessageID = c.Param("id"); return h.svc.EditMessage(uid(c), in) })
}
func (h *Handler) deleteMessage(c *gin.Context) {
	var in dto.DeleteMessageDTO
	bind(c, &in, func() (any, error) { in.MessageID = c.Param("id"); return h.svc.DeleteMessage(uid(c), in) })
}
func (h *Handler) createChatRoom(c *gin.Context) {
	var in dto.CreateChatRoomDTO
	bind(c, &in, func() (any, error) {
		if in.Type == "PRIVATE" {
			return h.svc.CreatePrivateChat(uid(c), in.ParticipantID)
		}
		return h.svc.CreateGroupChat(uid(c), in.Name, in.ParticipantIDs)
	})
}
func (h *Handler) getUserByEmail(c *gin.Context) {
	d, e := h.svc.GetUserByEmail(c.Param("email"))
	respond(c, d, e)
}
func bind(c *gin.Context, in any, f func() (any, error)) {
	if c.ShouldBindJSON(in) != nil {
		c.JSON(400, gin.H{"statusCode": 400, "message": "Validation failed"})
		return
	}
	d, e := f()
	respond(c, d, e)
}
func respond(c *gin.Context, d any, e error) {
	if e != nil {
		c.JSON(apperr.Status(e), gin.H{"statusCode": apperr.Status(e), "message": apperr.Message(e)})
		return
	}
	c.JSON(http.StatusOK, d)
}
