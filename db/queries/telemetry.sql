-- name: DeviceGetByAPIKeyHash
-- params: api_key_hash text
SELECT jsonb_build_object(
    'id', id,
    'device_code', device_code,
    'name', name,
    'is_active', is_active,
    'last_heartbeat_at', last_heartbeat_at
) FROM devices
WHERE api_key_hash = $1 AND is_active = TRUE;

-- name: DeviceGetByCode
-- params: device_code text
SELECT jsonb_build_object(
    'id', id,
    'device_code', device_code,
    'name', name,
    'is_active', is_active,
    'last_heartbeat_at', last_heartbeat_at
) FROM devices WHERE device_code = $1;

-- name: DeviceCreateOrUpdate
-- params: device_code text, name text, api_key_hash text
INSERT INTO devices (device_code, name, api_key_hash, is_active)
VALUES ($1, $2, $3, TRUE)
ON CONFLICT (device_code) DO UPDATE SET
    name = excluded.name,
    api_key_hash = excluded.api_key_hash,
    updated_at = NOW()
RETURNING jsonb_build_object('id', id, 'device_code', device_code, 'name', name);

-- name: DeviceUpdateHeartbeat
-- params: id bigint
UPDATE devices SET last_heartbeat_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: TelemetryIngest
-- params: device_id bigint, silo_weight numeric, motor_rpm int, drum_temperature numeric, drum_humidity numeric, coil_status text, motor_status text
INSERT INTO telemetry_logs (
    device_id, silo_weight, motor_rpm, drum_temperature, drum_humidity, coil_status, motor_status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING jsonb_build_object(
    'id', id,
    'device_id', device_id,
    'silo_weight', silo_weight,
    'motor_rpm', motor_rpm,
    'drum_temperature', drum_temperature,
    'drum_humidity', drum_humidity,
    'coil_status', coil_status,
    'motor_status', motor_status,
    'recorded_at', recorded_at
);

-- name: TelemetryLatest
-- params: device_id bigint
SELECT jsonb_build_object(
    'id', t.id,
    'device_id', t.device_id,
    'device_code', d.device_code,
    'device_name', d.name,
    'silo_weight', t.silo_weight,
    'motor_rpm', t.motor_rpm,
    'drum_temperature', t.drum_temperature,
    'drum_humidity', t.drum_humidity,
    'coil_status', t.coil_status,
    'motor_status', t.motor_status,
    'recorded_at', t.recorded_at,
    'last_heartbeat_at', d.last_heartbeat_at
) FROM telemetry_logs t
JOIN devices d ON d.id = t.device_id
WHERE ($1 = 0 OR t.device_id = $1)
ORDER BY t.recorded_at DESC, t.id DESC
LIMIT 1;

-- name: TelemetryListHistory
-- params: cursor_id bigint, limit int, device_id bigint, start_time text, end_time text
SELECT jsonb_build_object(
    'id', t.id,
    'device_id', t.device_id,
    'device_code', d.device_code,
    'silo_weight', t.silo_weight,
    'motor_rpm', t.motor_rpm,
    'drum_temperature', t.drum_temperature,
    'drum_humidity', t.drum_humidity,
    'coil_status', t.coil_status,
    'motor_status', t.motor_status,
    'recorded_at', t.recorded_at
) FROM telemetry_logs t
JOIN devices d ON d.id = t.device_id
WHERE ($1 = 0 OR t.id < $1)
  AND ($3 = 0 OR t.device_id = $3)
  AND ($4 = '' OR t.recorded_at >= $4::timestamptz)
  AND ($5 = '' OR t.recorded_at <= $5::timestamptz)
ORDER BY t.id DESC
LIMIT $2;

-- name: TelemetryGetStats
-- params: device_id bigint, start_time text, end_time text
SELECT jsonb_build_object(
    'count', COUNT(*),
    'min_silo_weight', COALESCE(MIN(silo_weight), 0),
    'max_silo_weight', COALESCE(MAX(silo_weight), 0),
    'avg_silo_weight', COALESCE(ROUND(AVG(silo_weight), 2), 0),
    'min_motor_rpm', COALESCE(MIN(motor_rpm), 0),
    'max_motor_rpm', COALESCE(MAX(motor_rpm), 0),
    'avg_motor_rpm', COALESCE(ROUND(AVG(motor_rpm), 1), 0),
    'min_drum_temp', COALESCE(MIN(drum_temperature), 0),
    'max_drum_temp', COALESCE(MAX(drum_temperature), 0),
    'avg_drum_temp', COALESCE(ROUND(AVG(drum_temperature), 2), 0),
    'min_drum_humidity', COALESCE(MIN(drum_humidity), 0),
    'max_drum_humidity', COALESCE(MAX(drum_humidity), 0),
    'avg_drum_humidity', COALESCE(ROUND(AVG(drum_humidity), 2), 0)
) FROM telemetry_logs
WHERE ($1 = 0 OR device_id = $1)
  AND ($2 = '' OR recorded_at >= $2::timestamptz)
  AND ($3 = '' OR recorded_at <= $3::timestamptz);
