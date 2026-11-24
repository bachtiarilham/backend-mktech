-- 004_telemetry_and_devices.up.sql
CREATE TABLE IF NOT EXISTS devices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    device_code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    api_key_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_heartbeat_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_devices_code UNIQUE (device_code)
);

CREATE TABLE IF NOT EXISTS telemetry_logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    silo_weight NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    motor_rpm INTEGER NOT NULL DEFAULT 0,
    drum_temperature NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    drum_humidity NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    coil_status VARCHAR(20) NOT NULL DEFAULT 'STANDBY',
    motor_status VARCHAR(20) NOT NULL DEFAULT 'STOPPED',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry_logs (device_id, recorded_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_recorded_at ON telemetry_logs (recorded_at DESC);
