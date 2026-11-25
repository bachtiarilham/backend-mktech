package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"ptmuda.tech/backend/config"
	authDelivery "ptmuda.tech/backend/internal/auth/delivery"
	authRepo "ptmuda.tech/backend/internal/auth/repository"
	authService "ptmuda.tech/backend/internal/auth/service"
	authUC "ptmuda.tech/backend/internal/auth/usecase"
	cfDelivery "ptmuda.tech/backend/internal/cashflow/delivery"
	cfRepo "ptmuda.tech/backend/internal/cashflow/repository"
	cfUC "ptmuda.tech/backend/internal/cashflow/usecase"
	"ptmuda.tech/backend/internal/delivery"
	"ptmuda.tech/backend/internal/delivery/rest"
	"ptmuda.tech/backend/internal/infra/postgres"
	prodDelivery "ptmuda.tech/backend/internal/product/delivery"
	prodRepo "ptmuda.tech/backend/internal/product/repository"
	prodUC "ptmuda.tech/backend/internal/product/usecase"
	telDelivery "ptmuda.tech/backend/internal/telemetry/delivery"
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

	ar := authRepo.New(dbStore)
	authPass := authService.Passwords{}
	authLimit := authService.NewLimiter()
	c.Auth = authUC.New(ar, dbStore, authPass, authLimit)

	pr := prodRepo.New(dbStore)
	c.Product = prodUC.New(pr)

	cr := cfRepo.New(dbStore)
	c.Cashflow = cfUC.New(cr)

	tr := telRepo.New(dbStore)
	hub := telService.NewHub()
	c.Telemetry = telUC.New(tr, hub)

	c.Handlers = delivery.Handlers{
		Health: rest.NewHealthHandler(dbStore),
		Auth: authDelivery.NewHandler(c.Auth, authDelivery.Options{
			CookieName: cfg.CookieName,
			Secure:     cfg.CookieSecure,
		}),
		Product:        prodDelivery.NewHandler(c.Product),
		Cashflow:       cfDelivery.NewHandler(c.Cashflow),
		Telemetry:      telDelivery.NewHandler(c.Telemetry),
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
