package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"ptmuda.tech/backend/internal/apperror"
	coreModel "ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/delivery/rest/response"
	prodModel "ptmuda.tech/backend/internal/product/model"
	"ptmuda.tech/backend/internal/product/repository"
)

type UseCase struct {
	repo repository.Repository
}

func New(repo repository.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) Create(ctx context.Context, in prodModel.CreateProductRequest, createdBy int64) (*prodModel.Product, error) {
	sku := strings.ToUpper(strings.TrimSpace(in.SKU))
	name := strings.TrimSpace(in.Name)
	if sku == "" || name == "" {
		return nil, apperror.Invalid("SKU dan Nama Produk wajib diisi")
	}
	in.SKU = sku
	in.Name = name

	existing, err := u.repo.GetBySKU(ctx, sku)
	if err == nil && existing != nil {
		return nil, apperror.Public(apperror.ErrConflict, fmt.Sprintf("Produk dengan SKU '%s' sudah terdaftar", sku))
	} else if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	if in.Unit == "" {
		in.Unit = "PCS"
	}
	in.Unit = strings.ToUpper(strings.TrimSpace(in.Unit))

	if in.Status == "" {
		if in.Stock > 0 {
			in.Status = "ACTIVE"
		} else {
			in.Status = "OUT_OF_STOCK"
		}
	} else {
		in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	}

	rec, err := u.repo.Create(ctx, in, createdBy)
	if err != nil {
		return nil, err
	}

	return toProduct(rec), nil
}

func (u *UseCase) GetByID(ctx context.Context, id int64) (*prodModel.Product, error) {
	if id <= 0 {
		return nil, apperror.ErrNotFound
	}
	rec, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toProduct(rec), nil
}

func (u *UseCase) List(ctx context.Context, f prodModel.ProductFilter) (*response.CursorPage[prodModel.Product], error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 10
	}
	records, err := u.repo.List(ctx, prodModel.ProductFilter{
		Cursor:   f.Cursor,
		Limit:    f.Limit + 1,
		Category: f.Category,
		Status:   f.Status,
		Search:   f.Search,
	})
	if err != nil {
		return nil, err
	}

	hasMore := len(records) > f.Limit
	if hasMore {
		records = records[:f.Limit]
	}

	items := make([]prodModel.Product, 0, len(records))
	var nextCursor string
	for _, r := range records {
		p := toProduct(r)
		items = append(items, *p)
		nextCursor = strconv.FormatInt(p.ID, 10)
	}

	if !hasMore {
		nextCursor = ""
	}

	return &response.CursorPage[prodModel.Product]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (u *UseCase) Update(ctx context.Context, id int64, in prodModel.UpdateProductRequest) (*prodModel.Product, error) {
	if id <= 0 {
		return nil, apperror.ErrNotFound
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, apperror.Invalid("Nama Produk wajib diisi")
	}
	if in.Unit == "" {
		in.Unit = "PCS"
	}
	in.Unit = strings.ToUpper(strings.TrimSpace(in.Unit))

	if in.Status == "" {
		if in.Stock > 0 {
			in.Status = "ACTIVE"
		} else {
			in.Status = "OUT_OF_STOCK"
		}
	} else {
		in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	}

	rec, err := u.repo.Update(ctx, id, in)
	if err != nil {
		return nil, err
	}
	return toProduct(rec), nil
}

func (u *UseCase) UpdateStock(ctx context.Context, id int64, in prodModel.UpdateStockRequest) (*prodModel.Product, error) {
	if id <= 0 {
		return nil, apperror.ErrNotFound
	}
	if in.Status == "" {
		if in.Stock == 0 {
			in.Status = "OUT_OF_STOCK"
		} else {
			in.Status = "ACTIVE"
		}
	} else {
		in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	}

	_, err := u.repo.UpdateStock(ctx, id, in.Stock, in.Status)
	if err != nil {
		return nil, err
	}
	return u.GetByID(ctx, id)
}

func (u *UseCase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return apperror.ErrNotFound
	}
	return u.repo.Delete(ctx, id)
}

func toProduct(r coreModel.Record) *prodModel.Product {
	p := &prodModel.Product{
		ID:           int64(coreModel.Number(r, "id")),
		SKU:          coreModel.Text(r, "sku"),
		Name:         coreModel.Text(r, "name"),
		Description:  coreModel.Text(r, "description"),
		Category:     coreModel.Text(r, "category"),
		CostPrice:    coreModel.Float(r, "cost_price"),
		SellingPrice: coreModel.Float(r, "selling_price"),
		Stock:        coreModel.Number(r, "stock"),
		Unit:         coreModel.Text(r, "unit"),
		ImageURL:     coreModel.Text(r, "image_url"),
		Status:       coreModel.Text(r, "status"),
		CreatedAt:    coreModel.Text(r, "created_at"),
		UpdatedAt:    coreModel.Text(r, "updated_at"),
	}
	if cb := coreModel.Number(r, "created_by"); cb > 0 {
		v := int64(cb)
		p.CreatedBy = &v
	}
	return p
}
