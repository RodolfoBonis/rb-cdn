package middlewares

import (
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

var (
	tagApiKey   = "X-API-Key"
	validatorMu sync.RWMutex
	validator   APIKeyValidator
)

type APIKeyValidator interface {
	ValidateAPIKey(apiKey string, serviceId string) (keyGuardian.ApiKeyData, error)
}

func OverrideValidatorForTest(v APIKeyValidator) {
	validatorMu.Lock()
	defer validatorMu.Unlock()
	validator = v
}

func RestoreDefaultValidator() {
	validatorMu.Lock()
	defer validatorMu.Unlock()
	validator = nil
}

func ProtectWithApiKey(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(tagApiKey)

		if apiKey == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			logger.Log.Error("API Key is required")
			return
		}

		var configs keyGuardian.ApiKeyData
		var err error

		validatorMu.RLock()
		if validator != nil {
			configs, err = validator.ValidateAPIKey(apiKey, config.EnvServiceId())
		} else {
			configs, err = keyGuardian.ValidateAPIKey(apiKey, config.EnvServiceId())
		}
		validatorMu.RUnlock()

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
