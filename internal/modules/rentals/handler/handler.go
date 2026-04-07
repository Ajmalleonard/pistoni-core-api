package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/rentals/dto"
	"pistoni/core/internal/modules/rentals/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(s *service.Service) *Handler { return &Handler{svc: s} }
func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("", h.create)
	r.POST("/calculate-price", h.calculatePrice)
	r.GET("", h.findAll)
	r.GET("/user/:userId", h.findByUser)
	r.GET("/:id", h.findOne)
	r.PATCH("/:id", h.update)
	r.DELETE("/:id", h.remove)
}
func (h *Handler) create(c *gin.Context) {
	var in dto.CreateRentalDTO
	bind(c, &in, func() (any, error) { return h.svc.Create(in) })
}
func (h *Handler) calculatePrice(c *gin.Context) {
	var in dto.CalculateRentalPriceDTO
	bind(c, &in, func() (any, error) { return h.svc.CalculateRentalPrice(in) })
}
func (h *Handler) findAll(c *gin.Context) {
	skip, _ := strconv.Atoi(c.Query("skip"))
	take, _ := strconv.Atoi(c.Query("take"))
	d, e := h.svc.FindAll(skip, take, c.Query("userId"), c.Query("status"))
	respond(c, d, e)
}
func (h *Handler) findByUser(c *gin.Context) {
	d, e := h.svc.FindByUser(c.Param("userId"))
	respond(c, d, e)
}
func (h *Handler) findOne(c *gin.Context) { d, e := h.svc.FindOne(c.Param("id")); respond(c, d, e) }
func (h *Handler) update(c *gin.Context) {
	var patch map[string]any
	if c.ShouldBindJSON(&patch) != nil {
		c.JSON(400, gin.H{"statusCode": 400, "message": "Validation failed"})
		return
	}
	d, e := h.svc.Update(c.Param("id"), patch)
	respond(c, d, e)
}
func (h *Handler) remove(c *gin.Context) { d, e := h.svc.Remove(c.Param("id")); respond(c, d, e) }
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
