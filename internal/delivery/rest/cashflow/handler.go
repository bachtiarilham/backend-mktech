package cashflow

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	cfModel "ptmuda.tech/backend/internal/cashflow/model"
	"ptmuda.tech/backend/internal/cashflow/usecase"
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest/response"
)

type Handler struct {
	uc *usecase.UseCase
}

func NewHandler(uc *usecase.UseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	r.POST("", h.Create)
	r.GET("", h.List)
	r.GET("/summary", h.Summary)
	r.GET("/:id", h.GetByID)
	r.PUT("/:id", h.Update)
	r.DELETE("/:id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var in cfModel.CreateCashflowRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Payload cashflow tidak valid"))
		return
	}

	admin, _ := middleware.Admin(c.Request.Context())
	item, err := h.uc.Create(c.Request.Context(), in, admin.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Created(c, "Transaksi cashflow berhasil dicatat", item)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID cashflow tidak valid"))
		return
	}

	item, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Detail transaksi cashflow", item)
}

func (h *Handler) List(c *gin.Context) {
	var cursor int64
	if cur := c.Query("cursor"); cur != "" {
		if parsed, err := strconv.ParseInt(cur, 10, 64); err == nil {
			cursor = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var productID int64
	if pid := c.Query("product_id"); pid != "" {
		if parsed, err := strconv.ParseInt(pid, 10, 64); err == nil {
			productID = parsed
		}
	}

	filter := cfModel.CashflowFilter{
		Cursor:    cursor,
		Limit:     limit,
		Type:      c.Query("type"),
		Category:  c.Query("category"),
		ProductID: productID,
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	page, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Daftar transaksi cashflow", page)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID cashflow tidak valid"))
		return
	}

	var in cfModel.UpdateCashflowRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Payload update cashflow tidak valid"))
		return
	}

	item, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Transaksi cashflow berhasil diperbarui", item)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID cashflow tidak valid"))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Transaksi cashflow berhasil dihapus", nil)
}

func (h *Handler) Summary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := h.uc.Summary(c.Request.Context(), startDate, endDate)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Rekapitulasi cashflow", summary)
}
