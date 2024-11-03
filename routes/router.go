package routes

import (
	"github.com/RodolfoBonis/rb-cdn/core/health"
	mediaRoutes "github.com/RodolfoBonis/rb-cdn/features/media/routes"
	streamRoutes "github.com/RodolfoBonis/rb-cdn/features/stream/routes"
	uploadRoutes "github.com/RodolfoBonis/rb-cdn/features/upload/routes"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitializeRoutes(router *gin.Engine) {

	root := router.Group("/v1")

	root.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	health.InjectRoute(root)
	uploadRoutes.InjectRoutes(root)
	streamRoutes.InjectRoutes(root)
	mediaRoutes.InjectRoutes(root)
}
