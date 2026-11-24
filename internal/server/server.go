// Package server runs the HTTP process and coordinates graceful shutdown.
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"ptmuda.tech/backend/config"
	"ptmuda.tech/backend/internal/app"
	"ptmuda.tech/backend/internal/delivery"
	"ptmuda.tech/backend/internal/infra/logger"
)

func Run(ctx context.Context, cfg config.Config) error {
	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		return errors.New("cannot bind HTTP_ADDR: " + err.Error())
	}
	defer listener.Close()

	container, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer container.Close()

	if container.Database == nil {
		logger.Warn("PostgreSQL is not configured; database routes and readiness will return 503")
	}

	routerOpts := delivery.Options{
		AllowedOrigins: cfg.AllowedOrigins,
		TrustedProxies: cfg.TrustedProxies,
		RequestTimeout: cfg.RequestTimeout,
	}

	server := &http.Server{
		Handler:           delivery.Router(routerOpts, container.Handlers),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout + time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32768,
	}

	stopped := make(chan error, 1)
	go func() {
		stopped <- server.Serve(listener)
	}()

	logger.Info("PT Muda Kreasi Teknologi Backend listening",
		zap.String("address", cfg.HTTPAddress),
	)

	select {
	case err := <-stopped:
		if !errors.Is(err, http.ErrServerClosed) {
			_ = server.Close()
			return errors.New("HTTP server error: " + err.Error())
		}
		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return errors.New("HTTP shutdown timed out: " + err.Error())
	}

	logger.Info("HTTP server stopped gracefully")
	return nil
}
