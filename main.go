package main

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/middlewares"
	"github.com/RodolfoBonis/rb-cdn/docs"
	"github.com/RodolfoBonis/rb-cdn/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	app := gin.New()

	err := app.SetTrustedProxies([]string{})

	if err != nil {
		appError := errors.RootError(err.Error())
		logger.Log.Error(appError.Message, appError.ToMap())
		panic(err)
	}

	config.SentryConfig()

	newRelicConfig := config.NewRelicConfig()

	middleware := middlewares.NewMonitoringMiddleware(newRelicConfig)

	app.Use(middleware.NewRelicMiddleware())
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

	routes.InitializeRoutes(app)

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

}

func init() {

	config.LoadEnvVars()

	logger.InitLogger()

	// Use this for open connection with DataBase
	//appError := services.OpenConnection()
	//
	//if appError != nil {
	//	logger.Log.Error(appError.Message, appError.ToMap())
	//	panic(appError)
	//}

	// Use this for Run Yours migrations
	// services.RunMigrations()

	// Use this for open connection with RabbitMQ
	// services.StartAmqpConnection()

	docs.SwaggerInfo.Title = "Rb CDN"
	docs.SwaggerInfo.Description = "This is a service for upload any media file to MINIO"

	versionFileName := "/version.txt"

	if config.EnvironmentConfig() == entities.Environment.Development {
		versionFileName = "version.txt"
	}

	version, err := os.ReadFile(versionFileName)
	if err != nil {
		docs.SwaggerInfo.Version = "unknown"
	} else {
		docs.SwaggerInfo.Version = strings.TrimSpace(string(version))
	}
	docs.SwaggerInfo.Host = "rb-cdn.rodolfodebonis.com.br"
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"https"}
}
