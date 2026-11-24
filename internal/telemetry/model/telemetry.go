package model

import "time"

type Device struct {
	ID              int64      `json:"id"`
	DeviceCode      string     `json:"device_code"`
	Name            string     `json:"name"`
	IsActive        bool       `json:"is_active"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
}

type IngestTelemetryRequest struct {
	DeviceCode      string  `json:"device_code" binding:"required"`
	SiloWeight      float64 `json:"silo_weight" binding:"gte=0"`
	MotorRPM        int     `json:"motor_rpm" binding:"gte=0"`
	DrumTemperature float64 `json:"drum_temperature"`
	DrumHumidity    float64 `json:"drum_humidity" binding:"gte=0,lte=100"`
	CoilStatus      string  `json:"coil_status"`  // ACTIVE | STANDBY | OFF | ERROR
	MotorStatus     string  `json:"motor_status"` // RUNNING | STOPPED | WARNING
}

type TelemetrySnapshot struct {
	ID              int64     `json:"id"`
	DeviceID        int64     `json:"device_id"`
	DeviceCode      string    `json:"device_code"`
	DeviceName      string    `json:"device_name"`
	SiloWeight      float64   `json:"silo_weight"`      // kg
	MotorRPM        int       `json:"motor_rpm"`        // RPM
	DrumTemperature float64   `json:"drum_temperature"` // °C
	DrumHumidity    float64   `json:"drum_humidity"`    // %
	CoilStatus      string    `json:"coil_status"`      // ACTIVE, STANDBY, OFF
	MotorStatus     string    `json:"motor_status"`     // RUNNING, STOPPED
	RecordedAt      time.Time `json:"recorded_at"`
	IsOnline        bool      `json:"is_online"`
}

type TelemetryHistoryFilter struct {
	Cursor    int64
	Limit     int
	DeviceID  int64
	StartTime string
	EndTime   string
}

type TelemetryStats struct {
	Count           int64   `json:"count"`
	MinSiloWeight   float64 `json:"min_silo_weight"`
	MaxSiloWeight   float64 `json:"max_silo_weight"`
	AvgSiloWeight   float64 `json:"avg_silo_weight"`
	MinMotorRPM     int     `json:"min_motor_rpm"`
	MaxMotorRPM     int     `json:"max_motor_rpm"`
	AvgMotorRPM     float64 `json:"avg_motor_rpm"`
	MinDrumTemp     float64 `json:"min_drum_temp"`
	MaxDrumTemp     float64 `json:"max_drum_temp"`
	AvgDrumTemp     float64 `json:"avg_drum_temp"`
	MinDrumHumidity float64 `json:"min_drum_humidity"`
	MaxDrumHumidity float64 `json:"max_drum_humidity"`
	AvgDrumHumidity float64 `json:"avg_drum_humidity"`
}
