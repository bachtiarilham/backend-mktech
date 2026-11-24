package rest

import (
	"context"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/delivery/rest/response"
)

type Readiness interface {
	Ready(context.Context) error
}

type HealthHandler struct {
	usecase Readiness
}

func NewHealthHandler(u Readiness) *HealthHandler {
	return &HealthHandler{usecase: u}
}

func (h *HealthHandler) Live(c *gin.Context) {
	response.Success(c, "Service is alive", gin.H{"status": "UP"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if h.usecase == nil {
		response.Success(c, "Service ready", gin.H{"status": "UP", "database": "NOT_CONFIGURED"})
		return
	}
	if err := h.usecase.Ready(c.Request.Context()); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "Service ready", gin.H{"status": "UP", "database": "CONNECTED"})
}
