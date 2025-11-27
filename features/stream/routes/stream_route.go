package routes

import (
	rbauth "github.com/RodolfoBonis/rb_auth_client"
	"github.com/RodolfoBonis/rb-cdn/features/stream/di"
	"github.com/gin-gonic/gin"
)

func InjectRoutes(route *gin.RouterGroup, authClient *rbauth.Client) {
	var uc = di.StreamInjection()

	streamRoute := route.Group("/stream")
	streamRoute.GET("/*objectPath", authClient.RequireAuth(), uc.StreamVideo)
}
