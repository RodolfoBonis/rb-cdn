package routes

import (
	"github.com/RodolfoBonis/rb-cdn/core/middlewares"
	"github.com/RodolfoBonis/rb-cdn/features/stream/di"
	"github.com/gin-gonic/gin"
)

func InjectRoutes(route *gin.RouterGroup) {
	var uc = di.StreamInjection()

	streamRoute := route.Group("/stream")
	streamRoute.GET("/*objectPath", middlewares.ProtectWithApiKey(uc.StreamVideo))
}
