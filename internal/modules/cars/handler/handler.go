package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/cars/dto"
	"pistoni/core/internal/modules/cars/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(s *service.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Register(r *gin.RouterGroup) {
	r.POST("", h.create)
	r.POST("/bulk", h.createBulk)
	r.GET("", h.findAll)
	r.GET("/types", h.types)
	r.GET("/status/:status", h.findByStatus)
	r.GET("/owner/:id", h.findByOwner)
	r.GET("/type/:vehicleType", h.findByVehicleType)
	r.GET("/:id", h.findOne)
	r.PATCH("/:id", h.update)
	r.DELETE("/:id", h.remove)
}

func (h *Handler) create(c *gin.Context) {
	var in dto.CreateCarDTO
	bind(c, &in, func() (any, error) { return h.svc.Create(in) })
}
func (h *Handler) createBulk(c *gin.Context) {
	var in dto.CreateBulkCarsDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateBulk(in.Cars) })
}
func (h *Handler) findAll(c *gin.Context) {
	skip, _ := strconv.Atoi(c.Query("skip"))
	take, _ := strconv.Atoi(c.Query("take"))
	status := c.Query("status")
	data, err := h.svc.FindAll(skip, take, status)
	if err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, data)
}
func (h *Handler) types(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"types": h.svc.CarTypes()}) }
func (h *Handler) findByStatus(c *gin.Context) {
	data, err := h.svc.FindByStatus(c.Param("status"))
	respond(c, data, err)
}
func (h *Handler) findByOwner(c *gin.Context) {
	data, err := h.svc.FindByOwner(c.Param("id"))
	respond(c, data, err)
}
func (h *Handler) findByVehicleType(c *gin.Context) {
	data, err := h.svc.FindByVehicleType(c.Param("vehicleType"))
	respond(c, data, err)
}
func (h *Handler) findOne(c *gin.Context) {
	data, err := h.svc.FindOne(c.Param("id"))
	respond(c, data, err)
}
func (h *Handler) update(c *gin.Context) {
	var patch map[string]any
	if c.ShouldBindJSON(&patch) != nil {
		c.JSON(400, gin.H{"statusCode": 400, "message": "Validation failed"})
		return
	}
	data, err := h.svc.Update(c.Param("id"), patch)
	respond(c, data, err)
}
func (h *Handler) remove(c *gin.Context) {
	data, err := h.svc.Remove(c.Param("id"))
	respond(c, data, err)
}

func bind(c *gin.Context, in any, f func() (any, error)) {
	if c.ShouldBindJSON(in) != nil {
		c.JSON(400, gin.H{"statusCode": 400, "message": "Validation failed"})
		return
	}
	data, err := f()
	respond(c, data, err)
}

func respond(c *gin.Context, data any, err error) {
	if err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, data)
}
