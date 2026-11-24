package repository

import (
	"context"

	cfModel "ptmuda.tech/backend/internal/cashflow/model"
	"ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/infra/postgres"
)

type Repository interface {
	Create(ctx context.Context, c cfModel.CreateCashflowRequest, createdBy int64) (model.Record, error)
	GetByID(ctx context.Context, id int64) (model.Record, error)
	List(ctx context.Context, f cfModel.CashflowFilter) ([]model.Record, error)
	Update(ctx context.Context, id int64, c cfModel.UpdateCashflowRequest) (model.Record, error)
	Delete(ctx context.Context, id int64) error
	Summary(ctx context.Context, startDate, endDate string) (model.Record, error)
}

type cashflowRepo struct {
	store *postgres.Store
}

func New(store *postgres.Store) Repository {
	return &cashflowRepo{store: store}
}

func (r *cashflowRepo) Create(ctx context.Context, c cfModel.CreateCashflowRequest, createdBy int64) (model.Record, error) {
	var prodID int64
	if c.ProductID != nil {
		prodID = *c.ProductID
	}
	return r.store.One(ctx, "CashflowCreate",
		c.Type, c.Category, c.Title, c.Description,
		c.Amount, prodID, c.TransactionDate, createdBy,
	)
}

func (r *cashflowRepo) GetByID(ctx context.Context, id int64) (model.Record, error) {
	return r.store.One(ctx, "CashflowGetByID", id)
}

func (r *cashflowRepo) List(ctx context.Context, f cfModel.CashflowFilter) ([]model.Record, error) {
	return r.store.Many(ctx, "CashflowList",
		f.Cursor, f.Limit, f.Type, f.Category, f.ProductID, f.StartDate, f.EndDate,
	)
}

func (r *cashflowRepo) Update(ctx context.Context, id int64, c cfModel.UpdateCashflowRequest) (model.Record, error) {
	var prodID int64
	if c.ProductID != nil {
		prodID = *c.ProductID
	}
	return r.store.One(ctx, "CashflowUpdate",
		id, c.Type, c.Category, c.Title, c.Description,
		c.Amount, prodID, c.TransactionDate,
	)
}

func (r *cashflowRepo) Delete(ctx context.Context, id int64) error {
	return r.store.Exec(ctx, "CashflowDelete", id)
}

func (r *cashflowRepo) Summary(ctx context.Context, startDate, endDate string) (model.Record, error) {
	return r.store.One(ctx, "CashflowSummary", startDate, endDate)
}
