package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"ptmuda.tech/backend/config"
	authRepo "ptmuda.tech/backend/internal/auth/repository"
	authService "ptmuda.tech/backend/internal/auth/service"
	authUC "ptmuda.tech/backend/internal/auth/usecase"
	cfRepo "ptmuda.tech/backend/internal/cashflow/repository"
	cfUC "ptmuda.tech/backend/internal/cashflow/usecase"
	"ptmuda.tech/backend/internal/delivery"
	"ptmuda.tech/backend/internal/delivery/rest"
	authHandler "ptmuda.tech/backend/internal/delivery/rest/auth"
	cfHandler "ptmuda.tech/backend/internal/delivery/rest/cashflow"
	prodHandler "ptmuda.tech/backend/internal/delivery/rest/product"
	telHandler "ptmuda.tech/backend/internal/delivery/rest/telemetry"
	"ptmuda.tech/backend/internal/infra/postgres"
	prodRepo "ptmuda.tech/backend/internal/product/repository"
	prodUC "ptmuda.tech/backend/internal/product/usecase"
	telRepo "ptmuda.tech/backend/internal/telemetry/repository"
	telService "ptmuda.tech/backend/internal/telemetry/service"
	telUC "ptmuda.tech/backend/internal/telemetry/usecase"
)

type Container struct {
	Database  *pgxpool.Pool
	Auth      *authUC.UseCase
	Product   *prodUC.UseCase
	Cashflow  *cfUC.UseCase
	Telemetry *telUC.UseCase
	Handlers  delivery.Handlers
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	c := &Container{}

	if cfg.DatabaseURL != "" {
		pool, err := postgres.Open(ctx, postgres.Options{
			URL:            cfg.DatabaseURL,
			MaxConns:       cfg.DBMaxConns,
			MinConns:       cfg.DBMinConns,
			ConnectTimeout: cfg.DBConnectTimeout,
		})
		if err != nil {
			return nil, err
		}
		c.Database = pool
	}

	dbStore := &postgres.Store{Pool: c.Database}

	// 1. Auth Module
	ar := authRepo.New(dbStore)
	authPass := authService.Passwords{}
	authLimit := authService.NewLimiter()
	c.Auth = authUC.New(ar, dbStore, authPass, authLimit)

	// 2. Product Module
	pr := prodRepo.New(dbStore)
	c.Product = prodUC.New(pr)

	// 3. Cashflow Module
	cr := cfRepo.New(dbStore)
	c.Cashflow = cfUC.New(cr)

	// 4. Telemetry Module
	tr := telRepo.New(dbStore)
	hub := telService.NewHub()
	c.Telemetry = telUC.New(tr, hub)

	// 5. Handlers
	c.Handlers = delivery.Handlers{
		Health: rest.NewHealthHandler(dbStore),
		Auth: authHandler.NewHandler(c.Auth, authHandler.Options{
			CookieName: cfg.CookieName,
			Secure:     cfg.CookieSecure,
		}),
		Product:        prodHandler.NewHandler(c.Product),
		Cashflow:       cfHandler.NewHandler(c.Cashflow),
		Telemetry:      telHandler.NewHandler(c.Telemetry),
		AdminVerifier:  c.Auth,
		DeviceVerifier: c.Telemetry,
	}

	return c, nil
}

func (c *Container) Close() {
	if c.Database != nil {
		c.Database.Close()
	}
}
