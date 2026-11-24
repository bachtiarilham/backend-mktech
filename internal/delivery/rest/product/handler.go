package product

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest/response"
	prodModel "ptmuda.tech/backend/internal/product/model"
	"ptmuda.tech/backend/internal/product/usecase"
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
	r.GET("/:id", h.GetByID)
	r.PUT("/:id", h.Update)
	r.PATCH("/:id/stock", h.UpdateStock)
	r.DELETE("/:id", h.Delete)
}

func (h *Handler) Create(c *gin.Context) {
	var in prodModel.CreateProductRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Payload produk tidak valid"))
		return
	}

	admin, _ := middleware.Admin(c.Request.Context())
	p, err := h.uc.Create(c.Request.Context(), in, admin.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Created(c, "Produk berhasil ditambahkan", p)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID produk tidak valid"))
		return
	}

	p, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Detail produk", p)
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

	filter := prodModel.ProductFilter{
		Cursor:   cursor,
		Limit:    limit,
		Category: c.Query("category"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	}

	page, err := h.uc.List(c.Request.Context(), filter)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Daftar produk", page)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID produk tidak valid"))
		return
	}

	var in prodModel.UpdateProductRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Payload update produk tidak valid"))
		return
	}

	p, err := h.uc.Update(c.Request.Context(), id, in)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Produk berhasil diperbarui", p)
}

func (h *Handler) UpdateStock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID produk tidak valid"))
		return
	}

	var in prodModel.UpdateStockRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.WriteError(c, apperror.Invalid("Payload update stok tidak valid"))
		return
	}

	p, err := h.uc.UpdateStock(c.Request.Context(), id, in)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Stok produk berhasil diperbarui", p)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(c, apperror.Invalid("ID produk tidak valid"))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		response.WriteError(c, err)
		return
	}

	response.Success(c, "Produk berhasil dihapus", nil)
}
