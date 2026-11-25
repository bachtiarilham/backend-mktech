package delivery

import (
	authDelivery "ptmuda.tech/backend/internal/auth/delivery"
	cfDelivery "ptmuda.tech/backend/internal/cashflow/delivery"
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest"
	prodDelivery "ptmuda.tech/backend/internal/product/delivery"
	telDelivery "ptmuda.tech/backend/internal/telemetry/delivery"
)

type Handlers struct {
	Health         *rest.HealthHandler
	Auth           *authDelivery.Handler
	Product        *prodDelivery.Handler
	Cashflow       *cfDelivery.Handler
	Telemetry      *telDelivery.Handler
	AdminVerifier  middleware.AdminVerifier
	DeviceVerifier middleware.DeviceVerifier
}
