package routes

import (
	"github.com/RodolfoBonis/rb-cdn/core/middlewares"
	"github.com/RodolfoBonis/rb-cdn/features/upload/di"
	"github.com/gin-gonic/gin"
)

func InjectRoutes(route *gin.RouterGroup) {
	var uc = di.UploadInjection()

	uploadRoute := route.Group("/upload")
	uploadRoute.POST("/", middlewares.ProtectWithApiKey(uc.Upload))
}
