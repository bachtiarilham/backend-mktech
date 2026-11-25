package delivery_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ptmuda.tech/backend/internal/app"
	"ptmuda.tech/backend/config"
	"ptmuda.tech/backend/internal/delivery"
)

func TestRouterRoutes(t *testing.T) {
	cfg := config.Config{
		HTTPAddress:    "127.0.0.1:8091",
		RequestTimeout: 5 * time.Second,
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	container, err := app.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
	defer container.Close()

	r := delivery.Router(delivery.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		RequestTimeout: cfg.RequestTimeout,
	}, container.Handlers)

	reqLive, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	wLive := httptest.NewRecorder()
	r.ServeHTTP(wLive, reqLive)
	if wLive.Code != http.StatusOK {
		t.Errorf("expected 200 on /healthz, got %d", wLive.Code)
	}

	reqProd, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/products", nil)
	wProd := httptest.NewRecorder()
	r.ServeHTTP(wProd, reqProd)
	if wProd.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on /api/v1/admin/products without token, got %d", wProd.Code)
	}

	reqIngest, _ := http.NewRequest(http.MethodPost, "/api/v1/telemetry/ingest", nil)
	wIngest := httptest.NewRecorder()
	r.ServeHTTP(wIngest, reqIngest)
	if wIngest.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on /api/v1/telemetry/ingest without X-Device-Key, got %d", wIngest.Code)
	}
}
