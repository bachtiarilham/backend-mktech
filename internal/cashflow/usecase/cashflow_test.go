package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"ptmuda.tech/backend/internal/apperror"
	cfModel "ptmuda.tech/backend/internal/cashflow/model"
	"ptmuda.tech/backend/internal/cashflow/usecase"
	coreModel "ptmuda.tech/backend/internal/core/model"
)

type mockCashflowRepo struct {
	summaryRecord coreModel.Record
}

func (m *mockCashflowRepo) Create(_ context.Context, c cfModel.CreateCashflowRequest, _ int64) (coreModel.Record, error) {
	return coreModel.Record{
		"id":               json.Number("1"),
		"type":             c.Type,
		"category":         c.Category,
		"title":            c.Title,
		"amount":           json.Number("500000"),
		"transaction_date": c.TransactionDate,
	}, nil
}

func (m *mockCashflowRepo) GetByID(_ context.Context, _ int64) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockCashflowRepo) List(_ context.Context, _ cfModel.CashflowFilter) ([]coreModel.Record, error) {
	return nil, nil
}

func (m *mockCashflowRepo) Update(_ context.Context, _ int64, _ cfModel.UpdateCashflowRequest) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockCashflowRepo) Delete(_ context.Context, _ int64) error {
	return nil
}

func (m *mockCashflowRepo) Summary(_ context.Context, _, _ string) (coreModel.Record, error) {
	return m.summaryRecord, nil
}

func TestCashflowUseCase_CreateValidation(t *testing.T) {
	uc := usecase.New(&mockCashflowRepo{})

	tests := []struct {
		name         string
		input        cfModel.CreateCashflowRequest
		wantErr      error
		expectedType string
	}{
		{
			name: "Accepts PROFIT lowercase and normalizes to uppercase",
			input: cfModel.CreateCashflowRequest{
				Type:            "profit",
				Category:        "penjualan",
				Title:           "Penjualan Unit Dryer",
				Amount:          500000,
				TransactionDate: "2026-11-25",
			},
			expectedType: "PROFIT",
		},
		{
			name: "Accepts LOSS and SAVING",
			input: cfModel.CreateCashflowRequest{
				Type:            "SAVING",
				Category:        "DANA_DARURAT",
				Title:           "Tabungan Usaha",
				Amount:          100000,
				TransactionDate: "2026-11-25",
			},
			expectedType: "SAVING",
		},
		{
			name: "Rejects invalid type",
			input: cfModel.CreateCashflowRequest{
				Type:            "BONUS",
				Category:        "LAINNYA",
				Title:           "Invalid Type",
				Amount:          10000,
				TransactionDate: "2026-11-25",
			},
			wantErr: apperror.ErrInvalid,
		},
		{
			name: "Rejects non-positive amount",
			input: cfModel.CreateCashflowRequest{
				Type:            "LOSS",
				Category:        "OPERASIONAL",
				Title:           "Zero Amount",
				Amount:          0,
				TransactionDate: "2026-11-25",
			},
			wantErr: apperror.ErrInvalid,
		},
		{
			name: "Rejects invalid date format",
			input: cfModel.CreateCashflowRequest{
				Type:            "LOSS",
				Category:        "OPERASIONAL",
				Title:           "Bad Date",
				Amount:          50000,
				TransactionDate: "25-11-2026",
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
			if got.Type != tt.expectedType {
				t.Errorf("expected Type %s, got %s", tt.expectedType, got.Type)
			}
		})
	}
}

func TestCashflowUseCase_Summary(t *testing.T) {
	uc := usecase.New(&mockCashflowRepo{
		summaryRecord: coreModel.Record{
			"total_profit": json.Number("15000000.00"),
			"total_loss":   json.Number("2500000.00"),
			"total_saving": json.Number("5000000.00"),
			"net_profit":   json.Number("12500000.00"),
		},
	})

	summary, err := uc.Summary(context.Background(), "2026-11-01", "2026-11-30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.NetProfit != 12500000.00 {
		t.Errorf("expected NetProfit 12500000.00, got %f", summary.NetProfit)
	}
}
