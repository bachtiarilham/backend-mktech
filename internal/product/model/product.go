package model

type Product struct {
	ID           int64   `json:"id"`
	SKU          string  `json:"sku"`
	Name         string  `json:"name"`
	Description  string  `json:"description,omitempty"`
	Category     string  `json:"category"`
	CostPrice    float64 `json:"cost_price"`
	SellingPrice float64 `json:"selling_price"`
	Stock        int     `json:"stock"`
	Unit         string  `json:"unit"`
	ImageURL     string  `json:"image_url,omitempty"`
	Status       string  `json:"status"` // ACTIVE, INACTIVE, OUT_OF_STOCK
	CreatedBy    *int64  `json:"created_by,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type CreateProductRequest struct {
	SKU          string  `json:"sku" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	Category     string  `json:"category" binding:"required"`
	CostPrice    float64 `json:"cost_price" binding:"gte=0"`
	SellingPrice float64 `json:"selling_price" binding:"gte=0"`
	Stock        int     `json:"stock" binding:"gte=0"`
	Unit         string  `json:"unit"`
	ImageURL     string  `json:"image_url"`
	Status       string  `json:"status"`
}

type UpdateProductRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	Category     string  `json:"category" binding:"required"`
	CostPrice    float64 `json:"cost_price" binding:"gte=0"`
	SellingPrice float64 `json:"selling_price" binding:"gte=0"`
	Stock        int     `json:"stock" binding:"gte=0"`
	Unit         string  `json:"unit"`
	ImageURL     string  `json:"image_url"`
	Status       string  `json:"status"`
}

type UpdateStockRequest struct {
	Stock  int    `json:"stock" binding:"gte=0"`
	Status string `json:"status"`
}

type ProductFilter struct {
	Cursor   int64
	Limit    int
	Category string
	Status   string
	Search   string
}
