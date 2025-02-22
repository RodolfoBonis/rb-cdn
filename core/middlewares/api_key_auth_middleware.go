package middlewares

import (
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

var tagApiKey = "X-Api-Key"

func ProtectWithApiKey(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			logger.Log.Error("API Key is required")
			return
		}

		if len(apiKey) < 1 {
			appError := errors.MiddlewareError("API Key is required")
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			return
		}

		configs, err := keyGuardian.ValidateAPIKey(apiKey, config.EnvServiceId())

		if err != nil {
			appError := errors.UnauthorizedError()
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			return
		}

		logger.Log.Info("Application that made the request: " + configs.ID.String())

		c.Set("configs", configs)
		handler(c)
	}
}
