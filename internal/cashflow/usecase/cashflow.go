package usecase

import (
	"context"
	"strconv"
	"strings"
	"time"

	"ptmuda.tech/backend/internal/apperror"
	cfModel "ptmuda.tech/backend/internal/cashflow/model"
	"ptmuda.tech/backend/internal/cashflow/repository"
	coreModel "ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/delivery/rest/response"
)

type UseCase struct {
	repo repository.Repository
}

func New(repo repository.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func validateType(t string) (string, error) {
	upper := strings.ToUpper(strings.TrimSpace(t))
	switch upper {
	case "PROFIT", "LOSS", "SAVING":
		return upper, nil
	default:
		return "", apperror.Invalid("Tipe cashflow harus salah satu dari: PROFIT, LOSS, SAVING")
	}
}

func validateDate(d string) (string, error) {
	trimmed := strings.TrimSpace(d)
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", apperror.Invalid("Format transaction_date harus YYYY-MM-DD")
	}
	return trimmed, nil
}

func (u *UseCase) Create(ctx context.Context, in cfModel.CreateCashflowRequest, createdBy int64) (*cfModel.Cashflow, error) {
	typ, err := validateType(in.Type)
	if err != nil {
		return nil, err
	}
	in.Type = typ

	date, err := validateDate(in.TransactionDate)
	if err != nil {
		return nil, err
	}
	in.TransactionDate = date

	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return nil, apperror.Invalid("Judul transaksi wajib diisi")
	}

	in.Category = strings.ToUpper(strings.TrimSpace(in.Category))
	if in.Category == "" {
		return nil, apperror.Invalid("Kategori cashflow wajib diisi")
	}

	if in.Amount <= 0 {
		return nil, apperror.Invalid("Nominal cashflow harus lebih besar dari 0")
	}

	rec, err := u.repo.Create(ctx, in, createdBy)
	if err != nil {
		return nil, err
	}
	return toCashflow(rec), nil
}

func (u *UseCase) GetByID(ctx context.Context, id int64) (*cfModel.Cashflow, error) {
	if id <= 0 {
		return nil, apperror.ErrNotFound
	}
	rec, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toCashflow(rec), nil
}

func (u *UseCase) List(ctx context.Context, f cfModel.CashflowFilter) (*response.CursorPage[cfModel.Cashflow], error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 10
	}
	if f.Type != "" {
		f.Type = strings.ToUpper(strings.TrimSpace(f.Type))
	}
	if f.Category != "" {
		f.Category = strings.ToUpper(strings.TrimSpace(f.Category))
	}

	records, err := u.repo.List(ctx, cfModel.CashflowFilter{
		Cursor:    f.Cursor,
		Limit:     f.Limit + 1,
		Type:      f.Type,
		Category:  f.Category,
		ProductID: f.ProductID,
		StartDate: f.StartDate,
		EndDate:   f.EndDate,
	})
	if err != nil {
		return nil, err
	}

	hasMore := len(records) > f.Limit
	if hasMore {
		records = records[:f.Limit]
	}

	items := make([]cfModel.Cashflow, 0, len(records))
	var nextCursor string
	for _, r := range records {
		c := toCashflow(r)
		items = append(items, *c)
		nextCursor = strconv.FormatInt(c.ID, 10)
	}

	if !hasMore {
		nextCursor = ""
	}

	return &response.CursorPage[cfModel.Cashflow]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (u *UseCase) Update(ctx context.Context, id int64, in cfModel.UpdateCashflowRequest) (*cfModel.Cashflow, error) {
	if id <= 0 {
		return nil, apperror.ErrNotFound
	}

	typ, err := validateType(in.Type)
	if err != nil {
		return nil, err
	}
	in.Type = typ

	date, err := validateDate(in.TransactionDate)
	if err != nil {
		return nil, err
	}
	in.TransactionDate = date

	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return nil, apperror.Invalid("Judul transaksi wajib diisi")
	}

	in.Category = strings.ToUpper(strings.TrimSpace(in.Category))
	if in.Category == "" {
		return nil, apperror.Invalid("Kategori cashflow wajib diisi")
	}

	if in.Amount <= 0 {
		return nil, apperror.Invalid("Nominal cashflow harus lebih besar dari 0")
	}

	rec, err := u.repo.Update(ctx, id, in)
	if err != nil {
		return nil, err
	}
	return toCashflow(rec), nil
}

func (u *UseCase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return apperror.ErrNotFound
	}
	return u.repo.Delete(ctx, id)
}

func (u *UseCase) Summary(ctx context.Context, startDate, endDate string) (*cfModel.CashflowSummary, error) {
	rec, err := u.repo.Summary(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &cfModel.CashflowSummary{
		TotalProfit: coreModel.Float(rec, "total_profit"),
		TotalLoss:   coreModel.Float(rec, "total_loss"),
		TotalSaving: coreModel.Float(rec, "total_saving"),
		NetProfit:   coreModel.Float(rec, "net_profit"),
	}, nil
}

func toCashflow(r coreModel.Record) *cfModel.Cashflow {
	c := &cfModel.Cashflow{
		ID:              int64(coreModel.Number(r, "id")),
		Type:            coreModel.Text(r, "type"),
		Category:        coreModel.Text(r, "category"),
		Title:           coreModel.Text(r, "title"),
		Description:     coreModel.Text(r, "description"),
		Amount:          coreModel.Float(r, "amount"),
		ProductName:     coreModel.Text(r, "product_name"),
		TransactionDate: coreModel.Text(r, "transaction_date"),
		CreatedAt:       coreModel.Text(r, "created_at"),
		UpdatedAt:       coreModel.Text(r, "updated_at"),
	}

	if pid := coreModel.Number(r, "product_id"); pid > 0 {
		v := int64(pid)
		c.ProductID = &v
	}
	if cb := coreModel.Number(r, "created_by"); cb > 0 {
		v := int64(cb)
		c.CreatedBy = &v
	}

	return c
}
