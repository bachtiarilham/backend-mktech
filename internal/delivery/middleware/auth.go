package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/auth/model"
	"ptmuda.tech/backend/internal/delivery/rest/response"
)

type AdminVerifier interface {
	AuthenticateAdmin(context.Context, string) (model.AdminIdentity, error)
}

type adminKey struct{}

func Admin(ctx context.Context) (model.AdminIdentity, bool) {
	v, ok := ctx.Value(adminKey{}).(model.AdminIdentity)
	return v, ok
}

func Bearer(c *gin.Context) (string, bool) {
	if len(c.Request.Header.Values("Authorization")) != 1 {
		return "", false
	}
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || len(parts[1]) > 8192 {
		return "", false
	}
	return parts[1], true
}

func RequireAdmin(v AdminVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := Bearer(c)
		if !ok {
			response.WriteError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}
		if v == nil {
			response.WriteError(c, apperror.ErrUnavailable)
			c.Abort()
			return
		}
		id, err := v.AuthenticateAdmin(c.Request.Context(), token)
		if err != nil {
			response.WriteError(c, err)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), adminKey{}, id))
		c.Next()
	}
}
