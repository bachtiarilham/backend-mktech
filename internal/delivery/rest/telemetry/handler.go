package telemetry

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest/response"
	telModel "ptmuda.tech/backend/internal/telemetry/model"
	"ptmuda.tech/backend/internal/telemetry/usecase"
)

type Handler struct {
	uc *usecase.UseCase
}

func NewHandler(uc *usecase.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) IngestRoute(r *gin.RouterGroup, devAuth gin.HandlerFunc) {
	r.POST("/ingest", devAuth, h.Ingest)
}

func (h *Handler) PublicRoutes(r *gin.RouterGroup) {
	r.GET("/stream", h.Stream)
	r.GET("/latest", h.Latest)
	r.GET("/history", h.History)
	r.GET("/stats", h.Stats)
}

func (h *Handler) Ingest(c *gin.Context) {
	dev, ok := middleware.Device(c.Request.Context())
	if !ok || dev == nil {
		response.WriteError(c, apperror.ErrUnauthorized)
		return
	}

	var in telModel.IngestTelemetryRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Payload telemetri tidak valid"))
		return
	}

	snapshot, err := h.uc.Ingest(c.Request.Context(), in, dev)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Created(c, "Data telemetri berhasil dicatat", snapshot)
}

func (h *Handler) Latest(c *gin.Context) {
	var deviceID int64
	if d := c.Query("device_id"); d != "" {
		if parsed, err := strconv.ParseInt(d, 10, 64); err == nil {
			deviceID = parsed
		}
	}

	snap, err := h.uc.Latest(c.Request.Context(), deviceID)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Data sensor terkini", snap)
}

func (h *Handler) History(c *gin.Context) {
	var cursor int64
	if cur := c.Query("cursor"); cur != "" {
		if parsed, err := strconv.ParseInt(cur, 10, 64); err == nil {
			cursor = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var deviceID int64
	if d := c.Query("device_id"); d != "" {
		if parsed, err := strconv.ParseInt(d, 10, 64); err == nil {
			deviceID = parsed
		}
	}

	page, err := h.uc.History(c.Request.Context(), telModel.TelemetryHistoryFilter{
		Cursor:    cursor,
		Limit:     limit,
		DeviceID:  deviceID,
		StartTime: c.Query("start_time"),
		EndTime:   c.Query("end_time"),
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Log data telemetri historis", page)
}

func (h *Handler) Stats(c *gin.Context) {
	var deviceID int64
	if d := c.Query("device_id"); d != "" {
		if parsed, err := strconv.ParseInt(d, 10, 64); err == nil {
			deviceID = parsed
		}
	}

	stats, err := h.uc.Stats(c.Request.Context(), deviceID, c.Query("start_time"), c.Query("end_time"))
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Statistik telemetri", stats)
}

func (h *Handler) Stream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ch := h.uc.Subscribe()
	defer h.uc.Unsubscribe(ch)

	// Send initial ping event
	c.SSEvent("ping", gin.H{"time": time.Now().Format(time.RFC3339)})
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case snap, ok := <-ch:
			if !ok {
				return false
			}
			b, err := json.Marshal(snap)
			if err != nil {
				return true
			}
			c.SSEvent("telemetry", string(b))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
