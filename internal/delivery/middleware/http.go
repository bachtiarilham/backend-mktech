// Package middleware owns transport limits, tracing, CORS and authentication extraction.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ptmuda.tech/backend/internal/delivery/rest/response"
	"ptmuda.tech/backend/internal/infra/logger"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if !requestIDPattern.MatchString(id) {
			var b [16]byte
			if _, e := rand.Read(b[:]); e != nil {
				response.WriteError(c, e)
				c.Abort()
				return
			}
			id = hex.EncodeToString(b[:])
		}
		c.Header("X-Request-ID", id)
		c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), id))
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		logger.FromContext(c.Request.Context()).Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("route", route),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.FromContext(c.Request.Context()).Error("HTTP panic recovered", zap.Any("panic", r))
				c.Abort()
				if !c.Writer.Written() {
					response.WriteError(c, nil)
				}
			}
		}()
		c.Next()
	}
}

func Limits(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := int64(2 * 1024 * 1024) // 2 MB
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func CORS(origins []string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[o] = true
	}
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		origin := c.GetHeader("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Device-Key")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
