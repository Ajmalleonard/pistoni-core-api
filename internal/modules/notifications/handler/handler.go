package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/notifications/dto"
	"pistoni/core/internal/modules/notifications/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(s *service.Service) *Handler { return &Handler{svc: s} }
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("", h.getNotifications)
	r.GET("/unread-count", h.getUnreadCount)
	r.POST("/mark-read", h.markAsRead)
	r.POST("/mark-all-read", h.markAllAsRead)
	r.DELETE("/:id", h.deleteNotification)
	r.POST("/test", h.createTestNotification)
}
func uid(c *gin.Context) string {
	if v, ok := c.Get("userID"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return "test-user-123"
}
func (h *Handler) getNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	d, e := h.svc.GetUserNotifications(uid(c), page, limit)
	respond(c, d, e)
}
func (h *Handler) getUnreadCount(c *gin.Context) {
	d, e := h.svc.GetUnreadCount(uid(c))
	respond(c, d, e)
}
func (h *Handler) markAsRead(c *gin.Context) {
	var in dto.MarkNotificationReadDTO
	bind(c, &in, func() (any, error) { return h.svc.MarkAsRead(in.NotificationIDs, uid(c)) })
}
func (h *Handler) markAllAsRead(c *gin.Context) {
	d, e := h.svc.MarkAllAsRead(uid(c))
	respond(c, d, e)
}
func (h *Handler) deleteNotification(c *gin.Context) {
	d, e := h.svc.DeleteNotification(c.Param("id"), uid(c))
	respond(c, d, e)
}
func (h *Handler) createTestNotification(c *gin.Context) {
	d, e := h.svc.CreateSystemNotification(uid(c), "Test Notification", "This is a test notification created from the API.")
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
