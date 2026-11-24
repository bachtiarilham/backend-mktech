package repository

import (
	"context"

	"ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/infra/postgres"
	telModel "ptmuda.tech/backend/internal/telemetry/model"
)

type Repository interface {
	GetDeviceByAPIKeyHash(ctx context.Context, hash string) (model.Record, error)
	GetDeviceByCode(ctx context.Context, code string) (model.Record, error)
	CreateDevice(ctx context.Context, code, name, keyHash string) (model.Record, error)
	UpdateDeviceHeartbeat(ctx context.Context, deviceID int64) error
	Ingest(ctx context.Context, in telModel.IngestTelemetryRequest, deviceID int64) (model.Record, error)
	Latest(ctx context.Context, deviceID int64) (model.Record, error)
	ListHistory(ctx context.Context, f telModel.TelemetryHistoryFilter) ([]model.Record, error)
	GetStats(ctx context.Context, deviceID int64, startTime, endTime string) (model.Record, error)
}

type telemetryRepo struct {
	store *postgres.Store
}

func New(store *postgres.Store) Repository {
	return &telemetryRepo{store: store}
}

func (r *telemetryRepo) GetDeviceByAPIKeyHash(ctx context.Context, hash string) (model.Record, error) {
	return r.store.One(ctx, "DeviceGetByAPIKeyHash", hash)
}

func (r *telemetryRepo) GetDeviceByCode(ctx context.Context, code string) (model.Record, error) {
	return r.store.One(ctx, "DeviceGetByCode", code)
}

func (r *telemetryRepo) CreateDevice(ctx context.Context, code, name, keyHash string) (model.Record, error) {
	return r.store.One(ctx, "DeviceCreateOrUpdate", code, name, keyHash)
}

func (r *telemetryRepo) UpdateDeviceHeartbeat(ctx context.Context, deviceID int64) error {
	return r.store.Exec(ctx, "DeviceUpdateHeartbeat", deviceID)
}

func (r *telemetryRepo) Ingest(ctx context.Context, in telModel.IngestTelemetryRequest, deviceID int64) (model.Record, error) {
	return r.store.One(ctx, "TelemetryIngest",
		deviceID, in.SiloWeight, in.MotorRPM, in.DrumTemperature, in.DrumHumidity,
		in.CoilStatus, in.MotorStatus,
	)
}

func (r *telemetryRepo) Latest(ctx context.Context, deviceID int64) (model.Record, error) {
	return r.store.One(ctx, "TelemetryLatest", deviceID)
}

func (r *telemetryRepo) ListHistory(ctx context.Context, f telModel.TelemetryHistoryFilter) ([]model.Record, error) {
	return r.store.Many(ctx, "TelemetryListHistory",
		f.Cursor, f.Limit, f.DeviceID, f.StartTime, f.EndTime,
	)
}

func (r *telemetryRepo) GetStats(ctx context.Context, deviceID int64, startTime, endTime string) (model.Record, error) {
	return r.store.One(ctx, "TelemetryGetStats", deviceID, startTime, endTime)
}
