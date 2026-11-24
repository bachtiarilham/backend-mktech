package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"ptmuda.tech/backend/internal/apperror"
	coreModel "ptmuda.tech/backend/internal/core/model"
	"ptmuda.tech/backend/internal/delivery/rest/response"
	telModel "ptmuda.tech/backend/internal/telemetry/model"
	"ptmuda.tech/backend/internal/telemetry/repository"
	"ptmuda.tech/backend/internal/telemetry/service"
)

type UseCase struct {
	repo repository.Repository
	hub  *service.Hub
}

func New(repo repository.Repository, hub *service.Hub) *UseCase {
	return &UseCase{
		repo: repo,
		hub:  hub,
	}
}

func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func (u *UseCase) AuthenticateDevice(ctx context.Context, apiKey string) (*telModel.Device, error) {
	trimmed := strings.TrimSpace(apiKey)
	if trimmed == "" {
		return nil, apperror.ErrUnauthorized
	}

	keyHash := HashAPIKey(trimmed)
	r, err := u.repo.GetDeviceByAPIKeyHash(ctx, keyHash)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	dev := &telModel.Device{
		ID:         int64(coreModel.Number(r, "id")),
		DeviceCode: coreModel.Text(r, "device_code"),
		Name:       coreModel.Text(r, "name"),
		IsActive:   coreModel.Flag(r, "is_active"),
	}

	if hb := coreModel.Text(r, "last_heartbeat_at"); hb != "" {
		if t, err := time.Parse(time.RFC3339, hb); err == nil {
			dev.LastHeartbeatAt = &t
		}
	}

	return dev, nil
}

func (u *UseCase) Ingest(ctx context.Context, in telModel.IngestTelemetryRequest, dev *telModel.Device) (*telModel.TelemetrySnapshot, error) {
	if in.CoilStatus == "" {
		in.CoilStatus = "STANDBY"
	} else {
		in.CoilStatus = strings.ToUpper(strings.TrimSpace(in.CoilStatus))
	}

	if in.MotorStatus == "" {
		if in.MotorRPM > 0 {
			in.MotorStatus = "RUNNING"
		} else {
			in.MotorStatus = "STOPPED"
		}
	} else {
		in.MotorStatus = strings.ToUpper(strings.TrimSpace(in.MotorStatus))
	}

	rec, err := u.repo.Ingest(ctx, in, dev.ID)
	if err != nil {
		return nil, err
	}

	_ = u.repo.UpdateDeviceHeartbeat(ctx, dev.ID)

	snapshot := &telModel.TelemetrySnapshot{
		ID:              int64(coreModel.Number(rec, "id")),
		DeviceID:        dev.ID,
		DeviceCode:      dev.DeviceCode,
		DeviceName:      dev.Name,
		SiloWeight:      in.SiloWeight,
		MotorRPM:        in.MotorRPM,
		DrumTemperature: in.DrumTemperature,
		DrumHumidity:    in.DrumHumidity,
		CoilStatus:      in.CoilStatus,
		MotorStatus:     in.MotorStatus,
		RecordedAt:      time.Now(),
		IsOnline:        true,
	}

	// Siarkan ke SSE Hub secara real-time
	u.hub.Broadcast(*snapshot)

	return snapshot, nil
}

func (u *UseCase) Latest(ctx context.Context, deviceID int64) (*telModel.TelemetrySnapshot, error) {
	rec, err := u.repo.Latest(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	snap := &telModel.TelemetrySnapshot{
		ID:              int64(coreModel.Number(rec, "id")),
		DeviceID:        int64(coreModel.Number(rec, "device_id")),
		DeviceCode:      coreModel.Text(rec, "device_code"),
		DeviceName:      coreModel.Text(rec, "device_name"),
		SiloWeight:      coreModel.Float(rec, "silo_weight"),
		MotorRPM:        coreModel.Number(rec, "motor_rpm"),
		DrumTemperature: coreModel.Float(rec, "drum_temperature"),
		DrumHumidity:    coreModel.Float(rec, "drum_humidity"),
		CoilStatus:      coreModel.Text(rec, "coil_status"),
		MotorStatus:     coreModel.Text(rec, "motor_status"),
		RecordedAt:      time.Now(),
		IsOnline:        true,
	}

	if rt := coreModel.Text(rec, "recorded_at"); rt != "" {
		if t, err := time.Parse(time.RFC3339, rt); err == nil {
			snap.RecordedAt = t
			// If last record is older than 30 seconds, consider offline
			snap.IsOnline = time.Since(t) < 30*time.Second
		}
	}

	return snap, nil
}

func (u *UseCase) History(ctx context.Context, f telModel.TelemetryHistoryFilter) (*response.CursorPage[telModel.TelemetrySnapshot], error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	records, err := u.repo.ListHistory(ctx, telModel.TelemetryHistoryFilter{
		Cursor:    f.Cursor,
		Limit:     f.Limit + 1,
		DeviceID:  f.DeviceID,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	})
	if err != nil {
		return nil, err
	}

	hasMore := len(records) > f.Limit
	if hasMore {
		records = records[:f.Limit]
	}

	items := make([]telModel.TelemetrySnapshot, 0, len(records))
	var nextCursor string
	for _, r := range records {
		snap := telModel.TelemetrySnapshot{
			ID:              int64(coreModel.Number(r, "id")),
			DeviceID:        int64(coreModel.Number(r, "device_id")),
			DeviceCode:      coreModel.Text(r, "device_code"),
			SiloWeight:      coreModel.Float(r, "silo_weight"),
			MotorRPM:        coreModel.Number(r, "motor_rpm"),
			DrumTemperature: coreModel.Float(r, "drum_temperature"),
			DrumHumidity:    coreModel.Float(r, "drum_humidity"),
			CoilStatus:      coreModel.Text(r, "coil_status"),
			MotorStatus:     coreModel.Text(r, "motor_status"),
		}
		if rt := coreModel.Text(r, "recorded_at"); rt != "" {
			if t, err := time.Parse(time.RFC3339, rt); err == nil {
				snap.RecordedAt = t
			}
		}
		items = append(items, snap)
		nextCursor = strconv.FormatInt(snap.ID, 10)
	}

	if !hasMore {
		nextCursor = ""
	}

	return &response.CursorPage[telModel.TelemetrySnapshot]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (u *UseCase) Stats(ctx context.Context, deviceID int64, startTime, endTime string) (*telModel.TelemetryStats, error) {
	r, err := u.repo.GetStats(ctx, deviceID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return &telModel.TelemetryStats{
		Count:           int64(coreModel.Number(r, "count")),
		MinSiloWeight:   coreModel.Float(r, "min_silo_weight"),
		MaxSiloWeight:   coreModel.Float(r, "max_silo_weight"),
		AvgSiloWeight:   coreModel.Float(r, "avg_silo_weight"),
		MinMotorRPM:     coreModel.Number(r, "min_motor_rpm"),
		MaxMotorRPM:     coreModel.Number(r, "max_motor_rpm"),
		AvgMotorRPM:     coreModel.Float(r, "avg_motor_rpm"),
		MinDrumTemp:     coreModel.Float(r, "min_drum_temp"),
		MaxDrumTemp:     coreModel.Float(r, "max_drum_temp"),
		AvgDrumTemp:     coreModel.Float(r, "avg_drum_temp"),
		MinDrumHumidity: coreModel.Float(r, "min_drum_humidity"),
		MaxDrumHumidity: coreModel.Float(r, "max_drum_humidity"),
		AvgDrumHumidity: coreModel.Float(r, "avg_drum_humidity"),
	}, nil
}

func (u *UseCase) RegisterDevice(ctx context.Context, code, name, apiKey string) (coreModel.Record, error) {
	keyHash := HashAPIKey(apiKey)
	return u.repo.CreateDevice(ctx, code, name, keyHash)
}

func (u *UseCase) Subscribe() chan telModel.TelemetrySnapshot {
	return u.hub.Subscribe()
}

func (u *UseCase) Unsubscribe(ch chan telModel.TelemetrySnapshot) {
	u.hub.Unsubscribe(ch)
}
