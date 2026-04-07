package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/users/service"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
)

type Handler struct{ svc *service.Service }

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("", h.create)
	r.GET("", h.findAll)
	r.GET("/role/:role", h.findByRole)
	r.GET("/email/:email", h.findByEmail)
	r.GET("/:id", h.findOne)
	r.PATCH("/:id", h.update)
	r.DELETE("/:id", h.remove)
}

func (h *Handler) create(c *gin.Context) {
	var in model.User
	if c.ShouldBindJSON(&in) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": 400, "message": "Validation failed"})
		return
	}
	if in.Role == "" {
		in.Role = model.UserRoleClient
	}
	if in.AuthProvider == "" {
		in.AuthProvider = model.AuthProviderEmail
	}
	u, err := h.svc.Create(in)
	handle(c, err, u)
}

func (h *Handler) findAll(c *gin.Context) {
	skip, _ := strconv.Atoi(c.Query("skip"))
	take, _ := strconv.Atoi(c.Query("take"))
	var role *model.UserRole
	if r := c.Query("role"); r != "" {
		rr := model.UserRole(r)
		role = &rr
	}
	data, err := h.svc.FindAll(skip, take, role)
	handle(c, err, data)
}
func (h *Handler) findByRole(c *gin.Context) {
	data, err := h.svc.FindByRole(model.UserRole(c.Param("role")))
	handle(c, err, data)
}
func (h *Handler) findByEmail(c *gin.Context) {
	data, err := h.svc.FindByEmail(c.Param("email"))
	handle(c, err, data)
}
func (h *Handler) findOne(c *gin.Context) {
	data, err := h.svc.FindOne(c.Param("id"))
	handle(c, err, data)
}

func (h *Handler) update(c *gin.Context) {
	var patch map[string]any
	if c.ShouldBindJSON(&patch) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": 400, "message": "Validation failed"})
		return
	}
	data, err := h.svc.Update(c.Param("id"), patch)
	handle(c, err, data)
}
func (h *Handler) remove(c *gin.Context) {
	data, err := h.svc.Remove(c.Param("id"))
	handle(c, err, data)
}

func handle(c *gin.Context, err error, data any) {
	if err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, data)
}
