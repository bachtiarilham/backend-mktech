package delivery

import (
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest"
	authHandler "ptmuda.tech/backend/internal/delivery/rest/auth"
	cfHandler "ptmuda.tech/backend/internal/delivery/rest/cashflow"
	prodHandler "ptmuda.tech/backend/internal/delivery/rest/product"
	telHandler "ptmuda.tech/backend/internal/delivery/rest/telemetry"
)

type Handlers struct {
	Health         *rest.HealthHandler
	Auth           *authHandler.Handler
	Product        *prodHandler.Handler
	Cashflow       *cfHandler.Handler
	Telemetry      *telHandler.Handler
	AdminVerifier  middleware.AdminVerifier
	DeviceVerifier middleware.DeviceVerifier
}
