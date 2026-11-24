package repository

import (
	"context"

	"ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/infra/postgres"
	prodModel "ptmuda.tech/backend/internal/product/model"
)

type Repository interface {
	Create(ctx context.Context, p prodModel.CreateProductRequest, createdBy int64) (model.Record, error)
	GetByID(ctx context.Context, id int64) (model.Record, error)
	GetBySKU(ctx context.Context, sku string) (model.Record, error)
	List(ctx context.Context, f prodModel.ProductFilter) ([]model.Record, error)
	Update(ctx context.Context, id int64, p prodModel.UpdateProductRequest) (model.Record, error)
	UpdateStock(ctx context.Context, id int64, stock int, status string) (model.Record, error)
	Delete(ctx context.Context, id int64) error
}

type productRepo struct {
	store *postgres.Store
}

func New(store *postgres.Store) Repository {
	return &productRepo{store: store}
}

func (r *productRepo) Create(ctx context.Context, p prodModel.CreateProductRequest, createdBy int64) (model.Record, error) {
	return r.store.One(ctx, "ProductCreate",
		p.SKU, p.Name, p.Description, p.Category,
		p.CostPrice, p.SellingPrice, p.Stock, p.Unit,
		p.ImageURL, p.Status, createdBy,
	)
}

func (r *productRepo) GetByID(ctx context.Context, id int64) (model.Record, error) {
	return r.store.One(ctx, "ProductGetByID", id)
}

func (r *productRepo) GetBySKU(ctx context.Context, sku string) (model.Record, error) {
	return r.store.One(ctx, "ProductGetBySKU", sku)
}

func (r *productRepo) List(ctx context.Context, f prodModel.ProductFilter) ([]model.Record, error) {
	return r.store.Many(ctx, "ProductList", f.Cursor, f.Limit, f.Category, f.Status, f.Search)
}

func (r *productRepo) Update(ctx context.Context, id int64, p prodModel.UpdateProductRequest) (model.Record, error) {
	return r.store.One(ctx, "ProductUpdate",
		id, p.Name, p.Description, p.Category,
		p.CostPrice, p.SellingPrice, p.Stock, p.Unit,
		p.ImageURL, p.Status,
	)
}

func (r *productRepo) UpdateStock(ctx context.Context, id int64, stock int, status string) (model.Record, error) {
	return r.store.One(ctx, "ProductUpdateStock", id, stock, status)
}

func (r *productRepo) Delete(ctx context.Context, id int64) error {
	return r.store.Exec(ctx, "ProductDelete", id)
}
