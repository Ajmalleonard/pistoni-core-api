package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/paymentmethods/dto"
	"pistoni/core/internal/modules/paymentmethods/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(s *service.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("", h.getPaymentMethods)
	r.POST("/card", h.createCardPaymentMethod)
	r.POST("/mobile", h.createMobilePaymentMethod)
	r.PUT("/:id", h.updatePaymentMethod)
	r.DELETE("/:id", h.deletePaymentMethod)
	r.PUT(":id/default", h.setDefaultPaymentMethod)
}
func uid(c *gin.Context) string {
	v, _ := c.Get("userID")
	s, _ := v.(string)
	if s == "" {
		s = "test-user-123"
	}
	return s
}
func (h *Handler) getPaymentMethods(c *gin.Context) {
	data, err := h.svc.GetPaymentMethods(uid(c))
	respond(c, data, err)
}
func (h *Handler) createCardPaymentMethod(c *gin.Context) {
	var in dto.CreateCardPaymentMethodDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateCardPaymentMethod(uid(c), in) })
}
func (h *Handler) createMobilePaymentMethod(c *gin.Context) {
	var in dto.CreateMobilePaymentMethodDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateMobilePaymentMethod(uid(c), in) })
}
func (h *Handler) updatePaymentMethod(c *gin.Context) {
	var in dto.UpdatePaymentMethodDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdatePaymentMethod(uid(c), c.Param("id"), in) })
}
func (h *Handler) deletePaymentMethod(c *gin.Context) {
	err := h.svc.DeletePaymentMethod(uid(c), c.Param("id"))
	if err != nil {
		respond(c, nil, err)
		return
	}
	c.Status(http.StatusOK)
}
func (h *Handler) setDefaultPaymentMethod(c *gin.Context) {
	data, err := h.svc.SetDefaultPaymentMethod(uid(c), c.Param("id"))
	respond(c, data, err)
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
