package routes

import (
	rbauth "github.com/RodolfoBonis/rb_auth_client"
	"github.com/RodolfoBonis/rb-cdn/features/media/di"
	"github.com/gin-gonic/gin"
)

func InjectRoutes(route *gin.RouterGroup, authClient *rbauth.Client) {
	var uc = di.MediaInjection()

	mediaRoute := route.Group("/cdn")
	mediaRoute.GET("/:bucket/*objectPath", authClient.RequireAuth(), uc.Media)
}
