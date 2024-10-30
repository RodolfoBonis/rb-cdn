package health

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func InjectRoute(route *gin.RouterGroup) {
	route.GET("/health_check", healthCheck)
}

// healthCheck godoc
// @Summary Health Check
// @Schemes
// @Description Check if This service is healthy
// @Tags HealthCheck
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 409 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /health_check [get]
func healthCheck(context *gin.Context) {
	context.String(http.StatusOK, "This Service is Healthy")
}
