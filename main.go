package main

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/bootstrap"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/middlewares"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/RodolfoBonis/rb-cdn/docs"
	"github.com/RodolfoBonis/rb-cdn/routes"
	rbauth "github.com/RodolfoBonis/rb_auth_client"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
	"time"
)

var authClient *rbauth.Client

func main() {
	// Auth client + capability sync live in main() (not init) so
	// tests that call initializeApp() directly don't trip the
	// rbauth.NewClient validation panic when their env doesn't
	// carry the service-account secret. Production runs main()
	// with the real env file populated.
	initializeAuthAndSync()

	app := gin.New()

	err := app.SetTrustedProxies([]string{})

	if err != nil {
		appError := errors.RootError(err.Error())
		logger.Log.Error(appError.Message, appError.ToMap())
		panic(err)
	}

	config.SentryConfig()

	//newRelicConfig := config.NewRelicConfig()

	middleware := middlewares.NewMonitoringMiddleware(*logger.Log)

	//app.Use(middleware.NewRelicMiddleware())
	app.Use(middleware.SentryMiddleware())
	app.Use(middleware.LogMiddleware)

	app.Use(gin.Logger())
	app.Use(gin.Recovery())
	app.Use(gin.ErrorLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Range", "X-Api-Key"},
		ExposeHeaders:    []string{"Content-Range", "Content-Length"},
		AllowCredentials: true,
	}))

	routes.InitializeRoutes(app, authClient)

	runPort := fmt.Sprintf(":%s", config.EnvPort())

	server := &http.Server{
		Addr:         runPort,
		Handler:      app,
		ReadTimeout:  10 * time.Minute, // Ajuste conforme necessário
		WriteTimeout: 10 * time.Minute, // Ajuste conforme necessário
	}

	err = server.ListenAndServe()
	if err != nil {
		appError := errors.RootError(err.Error())
		logger.Log.Error(appError.Message, appError.ToMap())
		panic(err)
	}

	logger.Log.Info("Server started successfully", map[string]interface{}{})

}

func init() {
	initializeApp()
}

func initializeApp() {
	config.LoadEnvVars()
	logger.InitLogger()

	versionFileName := "version.txt"
	if config.EnvironmentConfig() == entities.Environment.Production {
		versionFileName = "/version.txt"
	}

	version := "unknown"
	if content, err := os.ReadFile(versionFileName); err == nil {
		version = strings.TrimSpace(string(content))
	}

	docs.SwaggerInfo.Title = "Rb CDN"
	docs.SwaggerInfo.Description = "This is a service for upload any media file to MINIO"
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = "rb-cdn.rodolfodebonis.com.br"
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"https"}
}

// initializeAuthAndSync builds the rbauth client from the live env
// and runs the boot-time capability sync against rb_management_api.
// Called from main() — not init() — so tests that exercise
// initializeApp() in isolation don't have to populate the SDK's
// required Config fields.
//
// When the service-account secret is unset (typical in unit tests
// that exercise main() to validate listener startup), we skip auth
// init with a loud warning. Production runs always provide the
// secret via env; this skip is a test-time convenience, not a
// fail-open path the operator can accidentally rely on.
func initializeAuthAndSync() {
	if config.EnvRBCDNClientSecret() == "" {
		logger.Log.Warning("RB_CDN_CLIENT_SECRET unset — skipping auth init and capability sync (test mode?)",
			map[string]interface{}{})
		return
	}

	authClient = rbauth.NewClient(rbauth.Config{
		ManagementAPIURL: config.EnvManagementAPIURL(),
		ClientID:         config.EnvRBCDNClientID(),
		ClientSecret:     config.EnvRBCDNClientSecret(),
		KeycloakURL:      config.EnvKeycloakHost(),
		Realm:            config.EnvKeycloakRealm(),
		CacheTTL:         5 * time.Minute,
		EnableLogging:    true,
	})

	logger.Log.Info("Auth client initialized", map[string]interface{}{})

	// Boot-time capability sync. Reconciles rb-cdn's declared
	// capabilities (read/write base + per-bucket scopes) with the
	// management API catalog. Fail-closed: if mgmt-api is down or
	// the service Identity isn't registered yet, we panic instead
	// of starting the listener and silently serving against a
	// stale catalog.
	bootstrap.SyncCapabilities(
		services.NewMinioService(),
		logger.Log,
		bootstrap.DefaultFatal(logger.Log),
	)
}

// GetAuthClient returns the global auth client instance
func GetAuthClient() *rbauth.Client {
	return authClient
}
