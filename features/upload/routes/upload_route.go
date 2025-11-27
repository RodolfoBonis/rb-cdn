package routes

import (
	rbauth "github.com/RodolfoBonis/rb_auth_client"
	"github.com/RodolfoBonis/rb-cdn/features/upload/di"
	"github.com/gin-gonic/gin"
)

func InjectRoutes(route *gin.RouterGroup, authClient *rbauth.Client) {
	var uc = di.UploadInjection()

	uploadRoute := route.Group("/upload")
	uploadRoute.POST("/", authClient.RequireAuth(), uc.Upload)
}
