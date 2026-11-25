package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"ptmuda.tech/backend/internal/apperror"
	coreModel "ptmuda.tech/backend/internal/core/model"
	prodModel "ptmuda.tech/backend/internal/product/model"
	"ptmuda.tech/backend/internal/product/usecase"
)

type mockProductRepo struct {
	bySKU   map[string]coreModel.Record
	records []coreModel.Record
}

func (m *mockProductRepo) Create(_ context.Context, p prodModel.CreateProductRequest, createdBy int64) (coreModel.Record, error) {
	rec := coreModel.Record{
		"id":            json.Number("1"),
		"sku":           p.SKU,
		"name":          p.Name,
		"category":      p.Category,
		"cost_price":    json.Number("10000"),
		"selling_price": json.Number("15000"),
		"stock":         json.Number("0"),
		"unit":          p.Unit,
		"status":        p.Status,
		"created_by":    json.Number("1"),
	}
	if p.Stock > 0 {
		rec["stock"] = json.Number("10")
	}
	return rec, nil
}

func (m *mockProductRepo) GetByID(_ context.Context, id int64) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockProductRepo) GetBySKU(_ context.Context, sku string) (coreModel.Record, error) {
	if r, ok := m.bySKU[sku]; ok {
		return r, nil
	}
	return nil, apperror.ErrNotFound
}

func (m *mockProductRepo) List(_ context.Context, f prodModel.ProductFilter) ([]coreModel.Record, error) {
	if f.Limit < len(m.records) {
		return m.records[:f.Limit], nil
	}
	return m.records, nil
}

func (m *mockProductRepo) Update(_ context.Context, _ int64, _ prodModel.UpdateProductRequest) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockProductRepo) UpdateStock(_ context.Context, _ int64, _ int, _ string) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockProductRepo) Delete(_ context.Context, _ int64) error {
	return nil
}

func TestProductUseCase_Create(t *testing.T) {
	repo := &mockProductRepo{
		bySKU: map[string]coreModel.Record{
			"EXISTING-01": {"id": json.Number("99"), "sku": "EXISTING-01"},
		},
	}
	uc := usecase.New(repo)

	tests := []struct {
		name           string
		input          prodModel.CreateProductRequest
		wantErr        error
		expectedSKU    string
		expectedStatus string
	}{
		{
			name: "Normalizes SKU and defaults status ACTIVE when stock > 0",
			input: prodModel.CreateProductRequest{
				SKU:          "  dryer-01 ",
				Name:         "Rotary Dryer",
				Category:     "MESIN",
				CostPrice:    10000,
				SellingPrice: 15000,
				Stock:        10,
			},
			expectedSKU:    "DRYER-01",
			expectedStatus: "ACTIVE",
		},
		{
			name: "Sets OUT_OF_STOCK automatically when stock == 0",
			input: prodModel.CreateProductRequest{
				SKU:      "DRYER-02",
				Name:     "Silo Feeder",
				Category: "MESIN",
				Stock:    0,
			},
			expectedSKU:    "DRYER-02",
			expectedStatus: "OUT_OF_STOCK",
		},
		{
			name: "Rejects duplicate SKU with ErrConflict",
			input: prodModel.CreateProductRequest{
				SKU:      "existing-01",
				Name:     "Duplicate",
				Category: "MESIN",
			},
			wantErr: apperror.ErrConflict,
		},
		{
			name: "Rejects empty SKU with ErrInvalid",
			input: prodModel.CreateProductRequest{
				SKU:      "   ",
				Name:     "No SKU",
				Category: "MESIN",
			},
			wantErr: apperror.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uc.Create(context.Background(), tt.input, 1)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.SKU != tt.expectedSKU {
				t.Errorf("expected SKU %s, got %s", tt.expectedSKU, got.SKU)
			}
			if got.Status != tt.expectedStatus {
				t.Errorf("expected Status %s, got %s", tt.expectedStatus, got.Status)
			}
		})
	}
}

func TestProductUseCase_ListPagination(t *testing.T) {
	repo := &mockProductRepo{
		records: []coreModel.Record{
			{"id": json.Number("10"), "sku": "A10", "name": "P10"},
			{"id": json.Number("9"), "sku": "A9", "name": "P9"},
			{"id": json.Number("8"), "sku": "A8", "name": "P8"},
		},
	}
	uc := usecase.New(repo)

	page, err := uc.List(context.Background(), prodModel.ProductFilter{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}
	if !page.HasMore || page.NextCursor != "9" {
		t.Errorf("expected HasMore=true and NextCursor=9, got HasMore=%v NextCursor=%s", page.HasMore, page.NextCursor)
	}
}
