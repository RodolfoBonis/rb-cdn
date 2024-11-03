package middlewares

import (
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/gin-gonic/gin"
)

var tagApiKey = "X-Api-Key"

func ProtectWithApiKey(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader(tagApiKey)

		if accessToken == "" {
			accessToken = c.Query(tagApiKey)
		}

		if len(accessToken) < 1 {
			appError := errors.MiddlewareError("API Key is required")
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			return
		}

		configs, err := keyGuardian.ValidateAPIKey(accessToken, config.EnvServiceId())

		if err != nil {
			appError := errors.MiddlewareError(err.Error())
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
