package routes

import (
	"github.com/RodolfoBonis/rb-cdn/features/media/di"
	"github.com/gin-gonic/gin"
)

func InjectRoutes(route *gin.RouterGroup) {
	var uc = di.MediaInjection()

	mediaRoute := route.Group("/cdn")
	mediaRoute.GET("/:bucket/*objectPath", uc.Media)
}
