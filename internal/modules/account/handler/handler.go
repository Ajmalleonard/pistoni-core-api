package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/account/dto"
	"pistoni/core/internal/modules/account/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup, authRequired gin.HandlerFunc) {
	secure := r.Group("", authRequired)
	secure.GET("", h.overview)
	secure.GET("/overview", h.overview)
	secure.GET("/settings", h.getSettings)
	secure.PATCH("/settings", h.updateSettings)
	secure.GET("/safety", h.getSafety)
	secure.PATCH("/safety", h.updateSafety)
	secure.GET("/privacy", h.getPrivacy)
	secure.PATCH("/privacy", h.updatePrivacy)
	secure.GET("/security", h.getSecurity)
	secure.PATCH("/security", h.updateSecurity)
	secure.GET("/trusted-contacts", h.listTrustedContacts)
	secure.POST("/trusted-contacts", h.createTrustedContact)
	secure.PATCH("/trusted-contacts/:id", h.updateTrustedContact)
	secure.DELETE("/trusted-contacts/:id", h.deleteTrustedContact)
	secure.GET("/family-members", h.listFamilyMembers)
	secure.POST("/family-members", h.createFamilyMember)
	secure.PATCH("/family-members/:id", h.updateFamilyMember)
	secure.DELETE("/family-members/:id", h.deleteFamilyMember)
	secure.GET("/trusted-devices", h.listTrustedDevices)
	secure.POST("/trusted-devices", h.createTrustedDevice)
	secure.PATCH("/trusted-devices/:id", h.updateTrustedDevice)
	secure.DELETE("/trusted-devices/:id", h.deleteTrustedDevice)
	secure.GET("/documents", h.listDocuments)
	secure.POST("/documents", h.createDocument)
	secure.PATCH("/documents/:id", h.updateDocument)
	secure.DELETE("/documents/:id", h.deleteDocument)
	secure.GET("/calendars", h.listCalendars)
	secure.POST("/calendars", h.createCalendar)
	secure.PATCH("/calendars/:id", h.updateCalendar)
	secure.DELETE("/calendars/:id", h.deleteCalendar)
	secure.DELETE("", h.deleteAccount)
}

func (h *Handler) overview(c *gin.Context) {
	data, err := h.svc.GetOverview(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) getSettings(c *gin.Context) {
	data, err := h.svc.GetSettings(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) updateSettings(c *gin.Context) {
	var in dto.UpdateSettingsDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateSettings(currentUserID(c), in) })
}

func (h *Handler) getSafety(c *gin.Context) {
	data, err := h.svc.GetSafety(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) updateSafety(c *gin.Context) {
	var in dto.UpdateSafetyDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateSafety(currentUserID(c), in) })
}

func (h *Handler) getPrivacy(c *gin.Context) {
	data, err := h.svc.GetPrivacy(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) updatePrivacy(c *gin.Context) {
	var in dto.UpdatePrivacyDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdatePrivacy(currentUserID(c), in) })
}

func (h *Handler) getSecurity(c *gin.Context) {
	data, err := h.svc.GetSecurity(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) updateSecurity(c *gin.Context) {
	var in dto.UpdateSecurityDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateSecurity(currentUserID(c), in) })
}

func (h *Handler) listTrustedContacts(c *gin.Context) {
	data, err := h.svc.ListTrustedContacts(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) createTrustedContact(c *gin.Context) {
	var in dto.CreateTrustedContactDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateTrustedContact(currentUserID(c), in) })
}

func (h *Handler) updateTrustedContact(c *gin.Context) {
	var in dto.UpdateTrustedContactDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateTrustedContact(currentUserID(c), c.Param("id"), in) })
}

func (h *Handler) deleteTrustedContact(c *gin.Context) {
	data, err := h.svc.DeleteTrustedContact(currentUserID(c), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listFamilyMembers(c *gin.Context) {
	data, err := h.svc.ListFamilyMembers(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) createFamilyMember(c *gin.Context) {
	var in dto.CreateFamilyMemberDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateFamilyMember(currentUserID(c), in) })
}

func (h *Handler) updateFamilyMember(c *gin.Context) {
	var in dto.UpdateFamilyMemberDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateFamilyMember(currentUserID(c), c.Param("id"), in) })
}

func (h *Handler) deleteFamilyMember(c *gin.Context) {
	data, err := h.svc.DeleteFamilyMember(currentUserID(c), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listTrustedDevices(c *gin.Context) {
	data, err := h.svc.ListTrustedDevices(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) createTrustedDevice(c *gin.Context) {
	var in dto.CreateTrustedDeviceDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateTrustedDevice(currentUserID(c), in) })
}

func (h *Handler) updateTrustedDevice(c *gin.Context) {
	var in dto.UpdateTrustedDeviceDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateTrustedDevice(currentUserID(c), c.Param("id"), in) })
}

func (h *Handler) deleteTrustedDevice(c *gin.Context) {
	data, err := h.svc.DeleteTrustedDevice(currentUserID(c), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listDocuments(c *gin.Context) {
	data, err := h.svc.ListDocuments(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) createDocument(c *gin.Context) {
	var in dto.CreatePersonalDocumentDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateDocument(currentUserID(c), in) })
}

func (h *Handler) updateDocument(c *gin.Context) {
	var in dto.UpdatePersonalDocumentDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateDocument(currentUserID(c), c.Param("id"), in) })
}

func (h *Handler) deleteDocument(c *gin.Context) {
	data, err := h.svc.DeleteDocument(currentUserID(c), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) listCalendars(c *gin.Context) {
	data, err := h.svc.ListCalendars(currentUserID(c))
	respond(c, data, err)
}

func (h *Handler) createCalendar(c *gin.Context) {
	var in dto.CreateCalendarConnectionDTO
	bind(c, &in, func() (any, error) { return h.svc.CreateCalendar(currentUserID(c), in) })
}

func (h *Handler) updateCalendar(c *gin.Context) {
	var in dto.UpdateCalendarConnectionDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateCalendar(currentUserID(c), c.Param("id"), in) })
}

func (h *Handler) deleteCalendar(c *gin.Context) {
	data, err := h.svc.DeleteCalendar(currentUserID(c), c.Param("id"))
	respond(c, data, err)
}

func (h *Handler) deleteAccount(c *gin.Context) {
	data, err := h.svc.DeleteAccount(currentUserID(c))
	respond(c, data, err)
}

func bind(c *gin.Context, in any, call func() (any, error)) {
	if err := c.ShouldBindJSON(in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Validation failed"})
		return
	}
	data, err := call()
	respond(c, data, err)
}

func respond(c *gin.Context, data any, err error) {
	if err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, data)
}

func currentUserID(c *gin.Context) string {
	if v, ok := c.Get("userID"); ok {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}
