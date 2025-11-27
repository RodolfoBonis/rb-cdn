package routes

import (
	rbauth "github.com/RodolfoBonis/rb_auth_client"
	"github.com/RodolfoBonis/rb-cdn/core/health"
	mediaRoutes "github.com/RodolfoBonis/rb-cdn/features/media/routes"
	streamRoutes "github.com/RodolfoBonis/rb-cdn/features/stream/routes"
	uploadRoutes "github.com/RodolfoBonis/rb-cdn/features/upload/routes"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitializeRoutes(router *gin.Engine, authClient *rbauth.Client) {

	root := router.Group("/v1")

	root.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	health.InjectRoute(root)
	uploadRoutes.InjectRoutes(root, authClient)
	streamRoutes.InjectRoutes(root, authClient)
	mediaRoutes.InjectRoutes(root, authClient)
}
