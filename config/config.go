// Package config loads environment configuration without initializing services.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddress                       string
	DatabaseURL                       string
	DBMaxConns, DBMinConns            int32
	DBConnectTimeout, DBHealthTimeout time.Duration
	RequestTimeout, ShutdownTimeout   time.Duration
	AllowedOrigins                    []string
	TrustedProxies                    []string
	FrontendURL, CookieName           string
	CookieSecure                      bool
	BootstrapAdminEmail               string
	BootstrapAdminPassword            string
	BootstrapAdminName                string
}

func Load(path string) (Config, error) {
	if err := godotenv.Load(path); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("cannot load environment file: %w", err)
	}

	c := Config{
		HTTPAddress: env("HTTP_ADDR", "127.0.0.1:8091"),
		DatabaseURL: strings.TrimSpace(env("DATABASE_URL", "")),
	}

	_, port, err := net.SplitHostPort(c.HTTPAddress)
	if err != nil {
		return c, fmt.Errorf("HTTP_ADDR must contain host:port")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return c, fmt.Errorf("HTTP_ADDR port must be 1..65535")
	}

	max, err := strconv.ParseInt(env("DB_MAX_CONNS", "10"), 10, 32)
	if err != nil || max < 1 || max > 1000 {
		return c, fmt.Errorf("DB_MAX_CONNS must be 1..1000")
	}
	min, err := strconv.ParseInt(env("DB_MIN_CONNS", "0"), 10, 32)
	if err != nil || min < 0 || min > max {
		return c, fmt.Errorf("DB_MIN_CONNS must be 0..DB_MAX_CONNS")
	}
	c.DBMaxConns, c.DBMinConns = int32(max), int32(min)

	for _, field := range []struct {
		key, fallback string
		target        *time.Duration
	}{
		{"DB_CONNECT_TIMEOUT", "5s", &c.DBConnectTimeout},
		{"DB_HEALTH_TIMEOUT", "2s", &c.DBHealthTimeout},
		{"HTTP_REQUEST_TIMEOUT", "15s", &c.RequestTimeout},
		{"SHUTDOWN_TIMEOUT", "10s", &c.ShutdownTimeout},
	} {
		value, err := time.ParseDuration(env(field.key, field.fallback))
		if err != nil || value < time.Second || value > time.Minute {
			return c, fmt.Errorf("%s must be 1s..1m", field.key)
		}
		*field.target = value
	}

	seen := map[string]bool{}
	for _, raw := range strings.Split(env("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:8091"), ",") {
		origin := strings.TrimSpace(raw)
		if origin == "" {
			continue
		}
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return c, fmt.Errorf("ALLOWED_ORIGINS must contain explicit HTTP(S) origins")
		}
		if !seen[origin] {
			c.AllowedOrigins = append(c.AllowedOrigins, origin)
			seen[origin] = true
		}
	}

	c.FrontendURL = strings.TrimRight(env("FRONTEND_PUBLIC_URL", "http://localhost:3000"), "/")
	c.CookieName = env("SESSION_COOKIE_NAME", "pmkt_refresh")
	c.CookieSecure = strings.EqualFold(env("COOKIE_SECURE", "false"), "true")

	for _, p := range strings.Split(env("TRUSTED_PROXIES", "127.0.0.1,::1"), ",") {
		proxy := strings.TrimSpace(p)
		if proxy != "" {
			c.TrustedProxies = append(c.TrustedProxies, proxy)
		}
	}

	return c, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
