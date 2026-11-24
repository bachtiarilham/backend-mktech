package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"ptmuda.tech/backend/config"
	"ptmuda.tech/backend/internal/auth/repository"
	"ptmuda.tech/backend/internal/auth/service"
	"ptmuda.tech/backend/internal/auth/usecase"
	"ptmuda.tech/backend/internal/infra/logger"
	"ptmuda.tech/backend/internal/infra/postgres"
)

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	flag.Parse()

	cfg, err := config.Load(*envFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	logger.Init()
	defer logger.Sync()

	email := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_EMAIL"))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	username := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_USERNAME"))
	name := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_NAME"))

	if email == "" || password == "" {
		fmt.Println("Usage: Set BOOTSTRAP_ADMIN_EMAIL, BOOTSTRAP_ADMIN_PASSWORD, BOOTSTRAP_ADMIN_USERNAME, BOOTSTRAP_ADMIN_NAME in environment.")
		fmt.Println("Example:")
		fmt.Println("  $env:BOOTSTRAP_ADMIN_EMAIL='admin@ptmuda.tech'")
		fmt.Println("  $env:BOOTSTRAP_ADMIN_USERNAME='superadmin'")
		fmt.Println("  $env:BOOTSTRAP_ADMIN_NAME='Super Administrator'")
		fmt.Println("  $env:BOOTSTRAP_ADMIN_PASSWORD='StrongPassword123!'")
		fmt.Println("  go run ./cmd/bootstrap-admin -env .env")
		os.Exit(1)
	}

	if username == "" {
		parts := strings.Split(email, "@")
		username = parts[0]
	}
	if name == "" {
		name = username
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.Open(ctx, postgres.Options{
		URL:            cfg.DatabaseURL,
		MaxConns:       cfg.DBMaxConns,
		MinConns:       cfg.DBMinConns,
		ConnectTimeout: cfg.DBConnectTimeout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := &postgres.Store{Pool: pool}
	repo := repository.New(store)
	uc := usecase.New(repo, store, service.Passwords{}, service.NewLimiter())

	rec, err := uc.BootstrapAdmin(ctx, username, email, password, name, "SUPER_ADMIN")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap admin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Admin account successfully bootstrapped: %v\n", rec)
}
