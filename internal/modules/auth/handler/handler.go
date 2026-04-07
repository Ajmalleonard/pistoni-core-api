package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/auth/dto"
	"pistoni/core/internal/modules/auth/service"
	"pistoni/core/internal/platform/apperr"
)

type Handler struct{ svc *service.Service }

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *gin.RouterGroup, authRequired gin.HandlerFunc) {
	r.POST("/login", h.login)
	r.POST("/google", h.googleOAuth)
	r.POST("/apple", h.appleOAuth)
	r.POST("/refresh", h.refresh)
	r.POST("/check-email", h.checkEmail)
	r.POST("/validate-email", h.validateEmail)
	r.POST("/complete-registration", h.completeRegistration)
	r.POST("/resend-code", h.resendCode)
	r.POST("/register", h.register)
	r.POST("/verify-code", h.verifyCode)
	r.GET("/verify-email/:token", h.verifyEmail)
	secure := r.Group("", authRequired)
	secure.POST("/logout", h.logout)
	secure.GET("/profile", h.profile)
	secure.PATCH("/profile", h.updateProfile)
	secure.POST("/change-password", h.changePassword)
}

func (h *Handler) login(c *gin.Context) {
	var in dto.LoginDTO
	bind(c, &in, func() (any, error) { return h.svc.Login(in) })
}
func (h *Handler) googleOAuth(c *gin.Context) {
	var in dto.GoogleOAuthDTO
	bind(c, &in, func() (any, error) { return h.svc.GoogleOAuth(in) })
}
func (h *Handler) appleOAuth(c *gin.Context) {
	var in dto.AppleOAuthDTO
	bind(c, &in, func() (any, error) { return h.svc.AppleOAuth(in) })
}
func (h *Handler) refresh(c *gin.Context) {
	var in dto.RefreshDTO
	bind(c, &in, func() (any, error) { return h.svc.Refresh(in.RefreshToken) })
}
func (h *Handler) checkEmail(c *gin.Context) {
	var in dto.CheckEmailDTO
	bind(c, &in, func() (any, error) {
		if err := h.svc.RequestEmailVerification(in.Email); err != nil {
			return nil, err
		}
		return gin.H{"message": "Verification code sent to email", "nextStep": "validate-email", "email": in.Email}, nil
	})
}
func (h *Handler) validateEmail(c *gin.Context) {
	var in dto.ValidateEmailDTO
	bind(c, &in, func() (any, error) {
		if err := h.svc.VerifyEmailCode(in.Email, in.Code); err != nil {
			return nil, err
		}
		return gin.H{"message": "Email validated successfully", "nextStep": "complete-registration", "email": in.Email}, nil
	})
}
func (h *Handler) completeRegistration(c *gin.Context) {
	var in dto.CompleteRegistrationDTO
	bind(c, &in, func() (any, error) { return h.svc.CompleteRegistration(in) })
}
func (h *Handler) resendCode(c *gin.Context) {
	var in dto.CheckEmailDTO
	bind(c, &in, func() (any, error) {
		if err := h.svc.RequestEmailVerification(in.Email); err != nil {
			return nil, err
		}
		return gin.H{"message": "Verification code sent to email"}, nil
	})
}
func (h *Handler) register(c *gin.Context) {
	var in dto.RegisterDTO
	bind(c, &in, func() (any, error) { return h.svc.Register(in) })
}
func (h *Handler) verifyCode(c *gin.Context) {
	var in dto.ValidateEmailDTO
	bind(c, &in, func() (any, error) {
		if err := h.svc.VerifyRegistrationCode(in.Email, in.Code); err != nil {
			return nil, err
		}
		return gin.H{"message": "Email verified successfully"}, nil
	})
}
func (h *Handler) verifyEmail(c *gin.Context) {
	err := h.svc.VerifyEmailToken(c.Param("token"))
	if err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
func (h *Handler) logout(c *gin.Context) {
	p := currentPrincipal(c)
	if err := h.svc.Logout(p.User.ID); err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
func (h *Handler) profile(c *gin.Context) { c.JSON(http.StatusOK, currentPrincipal(c).User) }
func (h *Handler) updateProfile(c *gin.Context) {
	var in dto.UpdateProfileDTO
	bind(c, &in, func() (any, error) { return h.svc.UpdateProfile(currentPrincipal(c).User.ID, in) })
}
func (h *Handler) changePassword(c *gin.Context) {
	var in dto.ChangePasswordDTO
	bind(c, &in, func() (any, error) {
		err := h.svc.ChangePassword(currentPrincipal(c).User.ID, in.CurrentPassword, in.NewPassword)
		if err != nil {
			return nil, err
		}
		return gin.H{"success": true, "message": "Password changed successfully"}, nil
	})
}

func bind(c *gin.Context, in any, call func() (any, error)) {
	if err := c.ShouldBindJSON(in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": "Validation failed"})
		return
	}
	out, err := call()
	if err != nil {
		c.JSON(apperr.Status(err), gin.H{"statusCode": apperr.Status(err), "message": apperr.Message(err)})
		return
	}
	c.JSON(http.StatusOK, out)
}

func currentPrincipal(c *gin.Context) service.Principal {
	v, _ := c.Get("currentUser")
	return v.(service.Principal)
}
