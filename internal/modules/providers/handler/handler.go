package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authsvc "pistoni/core/internal/modules/auth/service"
	"pistoni/core/internal/modules/providers/dto"
	providersvc "pistoni/core/internal/modules/providers/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *providersvc.Service }

func New(s *providersvc.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Register(r *gin.RouterGroup, authRequired gin.HandlerFunc, adminOnly gin.HandlerFunc, sensitiveQueue gin.HandlerFunc, retryBackoff gin.HandlerFunc) {
	drivers := r.Group("/drivers")
	drivers.POST("/register", authRequired, sensitiveQueue, h.registerDriver)
	drivers.GET("", h.findDrivers)
	drivers.GET("/status/updates", h.pollDriverStatusUpdates)
	drivers.GET("/me/profile", authRequired, h.getMyDriverProfile)
	drivers.GET("/me/status", authRequired, h.getMyDriverStatus)
	drivers.PATCH("/me/status", authRequired, sensitiveQueue, h.updateMyDriverStatus)
	drivers.PATCH("/me/availability", authRequired, sensitiveQueue, h.updateMyDriverAvailability)
	drivers.PATCH("/:id/verification", authRequired, adminOnly, sensitiveQueue, h.updateDriverVerification)
	drivers.GET("/:id/status", h.getDriverStatus)
	drivers.GET("/:id/reviews", h.listDriverReviews)
	drivers.POST("/:id/reviews", authRequired, sensitiveQueue, h.createDriverReview)
	drivers.GET("/:id/counters", authRequired, adminOnly, h.getDriverAssignmentCounters)
	drivers.PATCH("/:id/counters", authRequired, adminOnly, sensitiveQueue, h.updateDriverAssignmentCounters)
	drivers.GET("/:id", h.findDriverByID)

	bgs := r.Group("/body-guards")
	bgs.POST("/join", authRequired, retryBackoff, sensitiveQueue, h.joinBodyGuardQueue)
	bgs.GET("/queue", authRequired, adminOnly, h.listBodyGuardQueue)
	bgs.POST("/queue/:id/approve", authRequired, adminOnly, retryBackoff, sensitiveQueue, h.approveBodyGuardQueue)
	bgs.POST("/queue/:id/reject", authRequired, adminOnly, retryBackoff, sensitiveQueue, h.rejectBodyGuardQueue)
	bgs.POST("/register", authRequired, retryBackoff, sensitiveQueue, h.registerBodyGuard)
	bgs.GET("", h.findBodyGuards)
	bgs.GET("/me/profile", authRequired, h.getMyBodyGuardProfile)
	bgs.GET("/me/queue", authRequired, h.getMyBodyGuardQueue)
	bgs.GET("/me/status", authRequired, h.getMyBodyGuardStatus)
	bgs.PATCH("/me/status", authRequired, retryBackoff, sensitiveQueue, h.updateMyBodyGuardStatus)
	bgs.PATCH("/me/availability", authRequired, retryBackoff, sensitiveQueue, h.updateMyBodyGuardAvailability)
	bgs.PATCH("/:id/verification", authRequired, adminOnly, retryBackoff, sensitiveQueue, h.updateBodyGuardVerification)
	bgs.GET("/:id/status", h.getBodyGuardStatus)
	bgs.GET("/:id", h.findBodyGuardByID)
}

func (h *Handler) registerDriver(c *gin.Context) {
	var in dto.RegisterDriverDTO
	bind(c, &in, func() (any, error) { return h.svc.RegisterDriver(principal(c).User.ID, in) })
}
func (h *Handler) findDrivers(c *gin.Context) {
	var q dto.FindProviderQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.FindDrivers(q)
	respond(c, data, err)
}
func (h *Handler) getMyDriverProfile(c *gin.Context) {
	data, err := h.svc.GetMyDriverProfile(principal(c).User.ID)
	respond(c, data, err)
}
func (h *Handler) getMyDriverStatus(c *gin.Context) {
	data, err := h.svc.GetMyDriverStatus(principal(c).User.ID)
	respond(c, data, err)
}
func (h *Handler) updateMyDriverStatus(c *gin.Context) {
	var in dto.UpdateDriverStatusDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateMyDriverStatus(principal(c).User.ID, in) })
}
func (h *Handler) pollDriverStatusUpdates(c *gin.Context) {
	var q dto.DriverStatusPollQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.PollDriverStatusUpdates(q)
	respond(c, data, err)
}
func (h *Handler) getDriverStatus(c *gin.Context) {
	data, err := h.svc.GetDriverStatus(c.Param("id"))
	respond(c, data, err)
}
func (h *Handler) listDriverReviews(c *gin.Context) {
	data, err := h.svc.ListDriverReviews(c.Param("id"))
	respond(c, data, err)
}
func (h *Handler) createDriverReview(c *gin.Context) {
	var in dto.CreateDriverReviewDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateDriverReview(c.Param("id"), principal(c).User.ID, in) })
}
func (h *Handler) getDriverAssignmentCounters(c *gin.Context) {
	data, err := h.svc.GetDriverAssignmentCounters(c.Param("id"))
	respond(c, data, err)
}
func (h *Handler) updateDriverAssignmentCounters(c *gin.Context) {
	var in dto.DriverAssignmentDeltaDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateDriverAssignmentCounters(c.Param("id"), in) })
}
func (h *Handler) updateMyDriverAvailability(c *gin.Context) {
	var in dto.UpdateAvailabilityDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateMyDriverAvailability(principal(c).User.ID, in.IsAvailable) })
}
func (h *Handler) updateDriverVerification(c *gin.Context) {
	var in dto.UpdateVerificationDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateDriverVerification(c.Param("id"), in) })
}
func (h *Handler) findDriverByID(c *gin.Context) {
	data, err := h.svc.FindDriverByID(c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) joinBodyGuardQueue(c *gin.Context) {
	var in dto.JoinBodyGuardQueueDTO
	bind(c, &in, func() (any, error) { return h.svc.JoinBodyGuardQueue(principal(c).User.ID, in) })
}
func (h *Handler) listBodyGuardQueue(c *gin.Context) {
	data, err := h.svc.ListBodyGuardQueue(c.Query("status"))
	respond(c, data, err)
}
func (h *Handler) approveBodyGuardQueue(c *gin.Context) {
	data, err := h.svc.ApproveBodyGuardQueue(c.Param("id"), principal(c).User.ID)
	respond(c, data, err)
}
func (h *Handler) rejectBodyGuardQueue(c *gin.Context) {
	var in dto.ReviewBodyGuardQueueDTO
	if c.ShouldBindJSON(&in) != nil {
		in = dto.ReviewBodyGuardQueueDTO{}
	}
	data, err := h.svc.RejectBodyGuardQueue(c.Param("id"), principal(c).User.ID, in.Reason)
	respond(c, data, err)
}
func (h *Handler) registerBodyGuard(c *gin.Context) {
	var in dto.RegisterBodyGuardDTO
	bind(c, &in, func() (any, error) { return h.svc.RegisterBodyGuard(principal(c).User.ID, in) })
}
func (h *Handler) findBodyGuards(c *gin.Context) {
	var q dto.FindProviderQuery
	_ = c.ShouldBindQuery(&q)
	data, err := h.svc.FindBodyGuards(q)
	respond(c, data, err)
}
func (h *Handler) getMyBodyGuardProfile(c *gin.Context) {
	data, err := h.svc.GetMyBodyGuardProfile(principal(c).User.ID)
	respond(c, data, err)
}
func (h *Handler) getMyBodyGuardQueue(c *gin.Context) {
	data, err := h.svc.GetMyBodyGuardQueue(principal(c).User.ID)
	respond(c, data, err)
}
func (h *Handler) getMyBodyGuardStatus(c *gin.Context) {
	data, err := h.svc.GetMyBodyGuardStatus(principal(c).User.ID)
	respond(c, data, err)
}
func (h *Handler) updateMyBodyGuardStatus(c *gin.Context) {
	var in dto.UpdateBodyGuardStatusDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateMyBodyGuardStatus(principal(c).User.ID, in) })
}
func (h *Handler) getBodyGuardStatus(c *gin.Context) {
	data, err := h.svc.GetBodyGuardStatus(c.Param("id"))
	respond(c, data, err)
}
func (h *Handler) updateMyBodyGuardAvailability(c *gin.Context) {
	var in dto.UpdateAvailabilityDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateMyBodyGuardAvailability(principal(c).User.ID, in.IsAvailable) })
}
func (h *Handler) updateBodyGuardVerification(c *gin.Context) {
	var in dto.UpdateVerificationDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateBodyGuardVerification(c.Param("id"), in) })
}
func (h *Handler) findBodyGuardByID(c *gin.Context) {
	data, err := h.svc.FindBodyGuardByID(c.Param("id"))
	respond(c, data, err)
}

func principal(c *gin.Context) authsvc.Principal { return c.MustGet("currentUser").(authsvc.Principal) }

func bind(c *gin.Context, in any, f func() (any, error)) {
	if c.ShouldBindJSON(in) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": 400, "message": "Validation failed"})
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
