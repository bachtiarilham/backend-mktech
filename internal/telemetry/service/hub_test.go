package service_test

import (
	"testing"
	"time"

	"ptmuda.tech/backend/internal/telemetry/model"
	"ptmuda.tech/backend/internal/telemetry/service"
)

func TestHubBroadcast(t *testing.T) {
	hub := service.NewHub()
	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	snap := model.TelemetrySnapshot{
		DeviceCode:      "TEST_ARDUINO",
		SiloWeight:      500.5,
		MotorRPM:        1200,
		DrumTemperature: 80.0,
		DrumHumidity:    12.5,
		CoilStatus:      "ACTIVE",
		MotorStatus:     "RUNNING",
	}

	go hub.Broadcast(snap)

	select {
	case received := <-ch:
		if received.DeviceCode != "TEST_ARDUINO" {
			t.Errorf("expected device code TEST_ARDUINO, got %s", received.DeviceCode)
		}
		if received.DrumTemperature != 80.0 {
			t.Errorf("expected drum temperature 80.0, got %f", received.DrumTemperature)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for telemetry broadcast")
	}
}
