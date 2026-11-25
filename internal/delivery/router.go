package delivery

import (
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ptmuda.tech/backend/internal/apperror"
	"ptmuda.tech/backend/internal/delivery/middleware"
	"ptmuda.tech/backend/internal/delivery/rest/response"
)

type Options struct {
	AllowedOrigins []string
	TrustedProxies []string
	RequestTimeout time.Duration
}

func Router(opts Options, h Handlers) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.HandleMethodNotAllowed = true

	_ = r.SetTrustedProxies(opts.TrustedProxies)

	r.Use(
		middleware.SecurityHeaders(),
		middleware.AccessLog(),
		middleware.Recover(),
		middleware.Limits(opts.RequestTimeout),
		middleware.CORS(opts.AllowedOrigins),
	)

	r.GET("/healthz", h.Health.Live)
	r.HEAD("/healthz", h.Health.Live)
	r.GET("/readyz", h.Health.Ready)
	r.HEAD("/readyz", h.Health.Ready)

	apiV1 := r.Group("/api/v1")
	authPublic := apiV1.Group("/admin/auth")
	h.Auth.PublicRoutes(authPublic)

	adminProtected := apiV1.Group("/admin", middleware.RequireAdmin(h.AdminVerifier))
	authProtected := adminProtected.Group("/auth")
	h.Auth.ProtectedRoutes(authProtected)

	productsGroup := adminProtected.Group("/products")
	h.Product.Routes(productsGroup)

	cashflowsGroup := adminProtected.Group("/cashflows")
	h.Cashflow.Routes(cashflowsGroup)

	telemetryGroup := apiV1.Group("/telemetry")
	h.Telemetry.IngestRoute(telemetryGroup, middleware.RequireDeviceKey(h.DeviceVerifier))
	h.Telemetry.PublicRoutes(telemetryGroup)

	r.NoRoute(func(c *gin.Context) {
		response.WriteError(c, apperror.ErrNotFound)
	})

	r.NoMethod(func(c *gin.Context) {
		methods := []string{}
		for _, route := range r.Routes() {
			if route.Path == c.Request.URL.Path {
				methods = append(methods, route.Method)
			}
		}
		sort.Strings(methods)
		c.Header("Allow", strings.Join(methods, ", "))
		response.JSON(c, 405, false, "Method not allowed", nil, nil)
	})

	return r
}
