package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"ptmuda.tech/backend/config"
	"ptmuda.tech/backend/internal/infra/logger"
	"ptmuda.tech/backend/internal/server"
)

func main() {
	envPath := flag.String("env", ".env", "Path to environment file")
	flag.Parse()

	cfg, err := config.Load(*envPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger.Init()
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("Starting PT Muda Kreasi Teknologi Backend Server...",
		zap.String("version", "1.0.0"),
		zap.String("env", *envPath),
	)

	if err := server.Run(ctx, cfg); err != nil {
		logger.Error("Server terminated with error", zap.Error(err))
		os.Exit(1)
	}
}
