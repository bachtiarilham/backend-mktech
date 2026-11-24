package model

type Cashflow struct {
	ID              int64   `json:"id"`
	Type            string  `json:"type"` // PROFIT, LOSS, SAVING
	Category        string  `json:"category"`
	Title           string  `json:"title"`
	Description     string  `json:"description,omitempty"`
	Amount          float64 `json:"amount"`
	ProductID       *int64  `json:"product_id,omitempty"`
	ProductName     string  `json:"product_name,omitempty"`
	TransactionDate string  `json:"transaction_date"`
	CreatedBy       *int64  `json:"created_by,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type CreateCashflowRequest struct {
	Type            string  `json:"type" binding:"required"` // PROFIT, LOSS, SAVING
	Category        string  `json:"category" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount" binding:"gt=0"`
	ProductID       *int64  `json:"product_id"`
	TransactionDate string  `json:"transaction_date" binding:"required"` // YYYY-MM-DD
}

type UpdateCashflowRequest struct {
	Type            string  `json:"type" binding:"required"`
	Category        string  `json:"category" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount" binding:"gt=0"`
	ProductID       *int64  `json:"product_id"`
	TransactionDate string  `json:"transaction_date" binding:"required"`
}

type CashflowFilter struct {
	Cursor    int64
	Limit     int
	Type      string
	Category  string
	ProductID int64
	StartDate string
	EndDate   string
}

type CashflowSummary struct {
	TotalProfit float64 `json:"total_profit"`
	TotalLoss   float64 `json:"total_loss"`
	TotalSaving float64 `json:"total_saving"`
	NetProfit   float64 `json:"net_profit"`
}
