package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/auth/model"
	"ptmuda.tech/backend/internal/auth/usecase"
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest/response"
)

type Options struct {
	CookieName string
	Secure     bool
}

type Handler struct {
	uc   *usecase.UseCase
	opts Options
}

func NewHandler(uc *usecase.UseCase, opts Options) *Handler {
	return &Handler{uc: uc, opts: opts}
}

func (h *Handler) PublicRoutes(r *gin.RouterGroup) {
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
}

func (h *Handler) ProtectedRoutes(r *gin.RouterGroup) {
	r.POST("/logout", h.Logout)
	r.GET("/me", h.Me)
}

func (h *Handler) Login(c *gin.Context) {
	var in model.LoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Format request login tidak valid"))
		return
	}

	ip := c.ClientIP()
	agent := c.Request.UserAgent()

	resp, tokens, err := h.uc.Login(c.Request.Context(), in, ip, agent)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.opts.CookieName, tokens.RefreshToken, int(7*24*3600), "/", "", h.opts.Secure, true)

	response.Success(c, "Login admin berhasil", resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	cookie, err := c.Cookie(h.opts.CookieName)
	if err != nil || cookie == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if bErr := c.ShouldBindJSON(&body); bErr == nil && body.RefreshToken != "" {
			cookie = body.RefreshToken
		}
	}

	if cookie == "" {
		response.WriteError(c, apperror.ErrUnauthorized)
		return
	}

	ip := c.ClientIP()
	agent := c.Request.UserAgent()

	tokens, err := h.uc.Refresh(c.Request.Context(), cookie, ip, agent)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.opts.CookieName, tokens.RefreshToken, int(7*24*3600), "/", "", h.opts.Secure, true)

	response.Success(c, "Token berhasil diperbarui", gin.H{
		"access_token": tokens.AccessToken,
		"expires_in":   900,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	admin, ok := middleware.Admin(c.Request.Context())
	if ok && admin.SessionID > 0 {
		_ = h.uc.Logout(c.Request.Context(), admin.SessionID)
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.opts.CookieName, "", -1, "/", "", h.opts.Secure, true)

	response.Success(c, "Logout berhasil", nil)
}

func (h *Handler) Me(c *gin.Context) {
	admin, ok := middleware.Admin(c.Request.Context())
	if !ok {
		response.WriteError(c, apperror.ErrUnauthorized)
		return
	}

	profile, err := h.uc.Profile(c.Request.Context(), admin.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Profil admin", profile)
}
