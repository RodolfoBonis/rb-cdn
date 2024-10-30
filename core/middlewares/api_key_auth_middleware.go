package middlewares

import (
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/gin-gonic/gin"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
)

func ProtectWithApiKey(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("X-Api-Key")

		if len(accessToken) < 1 {
			appError := errors.MiddlewareError("API Key is required")
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		configs, err := keyGuardian.ValidateAPIKey(accessToken, config.EnvServiceId())

		if err != nil {
			appError := errors.MiddlewareError(err.Error())
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		logger.Log.Info("Application that made the request: " + configs.ID.String())

		c.Set("configs", configs)
		handler(c)
	}
}
