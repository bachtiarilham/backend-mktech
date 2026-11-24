package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/delivery/rest/response"
	telModel "ptmuda.tech/backend/internal/telemetry/model"
)

type DeviceVerifier interface {
	AuthenticateDevice(context.Context, string) (*telModel.Device, error)
}

type deviceKey struct{}

func Device(ctx context.Context) (*telModel.Device, bool) {
	v, ok := ctx.Value(deviceKey{}).(*telModel.Device)
	return v, ok
}

func RequireDeviceKey(v DeviceVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("X-Device-Key"))
		if key == "" {
			response.WriteError(c, apperror.Public(apperror.ErrUnauthorized, "Header X-Device-Key wajib disertakan"))
			c.Abort()
			return
		}

		if v == nil {
			response.WriteError(c, apperror.ErrUnavailable)
			c.Abort()
			return
		}

		dev, err := v.AuthenticateDevice(c.Request.Context(), key)
		if err != nil {
			response.WriteError(c, apperror.Public(apperror.ErrUnauthorized, "Device API key tidak valid atau device nonaktif"))
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), deviceKey{}, dev))
		c.Next()
	}
}
