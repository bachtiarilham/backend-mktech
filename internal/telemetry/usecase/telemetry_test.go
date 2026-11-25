package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"ptmuda.tech/backend/internal/apperror"
	coreModel "ptmuda.tech/backend/internal/core/model"
	telModel "ptmuda.tech/backend/internal/telemetry/model"
	"ptmuda.tech/backend/internal/telemetry/service"
	"ptmuda.tech/backend/internal/telemetry/usecase"
)

type mockTelemetryRepo struct {
	device coreModel.Record
}

func (m *mockTelemetryRepo) GetDeviceByAPIKeyHash(_ context.Context, hash string) (coreModel.Record, error) {
	if m.device != nil && hash == usecase.HashAPIKey("valid-key") {
		return m.device, nil
	}
	return nil, apperror.ErrNotFound
}

func (m *mockTelemetryRepo) GetDeviceByCode(_ context.Context, _ string) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockTelemetryRepo) CreateDevice(_ context.Context, _, _, _ string) (coreModel.Record, error) {
	return nil, nil
}

func (m *mockTelemetryRepo) UpdateDeviceHeartbeat(_ context.Context, _ int64) error {
	return nil
}

func (m *mockTelemetryRepo) Ingest(_ context.Context, _ telModel.IngestTelemetryRequest, _ int64) (coreModel.Record, error) {
	return coreModel.Record{"id": json.Number("10")}, nil
}

func (m *mockTelemetryRepo) Latest(_ context.Context, _ int64) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func (m *mockTelemetryRepo) ListHistory(_ context.Context, _ telModel.TelemetryHistoryFilter) ([]coreModel.Record, error) {
	return nil, nil
}

func (m *mockTelemetryRepo) GetStats(_ context.Context, _ int64, _, _ string) (coreModel.Record, error) {
	return nil, apperror.ErrNotFound
}

func TestTelemetryUseCase_IngestAndBroadcast(t *testing.T) {
	hub := service.NewHub()
	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	uc := usecase.New(&mockTelemetryRepo{}, hub)
	dev := &telModel.Device{ID: 1, DeviceCode: "ARDUINO_01", Name: "Dryer 1", IsActive: true}

	snap, err := uc.Ingest(context.Background(), telModel.IngestTelemetryRequest{
		DeviceCode:      "ARDUINO_01",
		SiloWeight:      850.5,
		MotorRPM:        1420,
		DrumTemperature: 76.4,
		DrumHumidity:    14.2,
		CoilStatus:      "active",
	}, dev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.CoilStatus != "ACTIVE" {
		t.Errorf("expected normalized CoilStatus ACTIVE, got %s", snap.CoilStatus)
	}
	if snap.MotorStatus != "RUNNING" {
		t.Errorf("expected auto MotorStatus RUNNING when RPM > 0, got %s", snap.MotorStatus)
	}

	select {
	case received := <-ch:
		if received.SiloWeight != 850.5 {
			t.Errorf("expected broadcasted SiloWeight 850.5, got %f", received.SiloWeight)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for telemetry SSE broadcast")
	}
}
