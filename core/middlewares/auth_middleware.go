package middlewares

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"strings"

	jsonToken "github.com/golang-jwt/jwt/v4"
)

func Protect(handler gin.HandlerFunc, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		keycloakDataAccess := config.EnvKeyCloak()
		authHeader := c.GetHeader("Authorization")

		if len(authHeader) < 1 {
			err := errors.InvalidTokenError()
			httpError := err.ToHttpError()
			logger.Log.Error(err.Message, err.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		accessToken := strings.Split(authHeader, " ")[1]

		rptResult, err := services.AuthClient.RetrospectToken(
			c,
			accessToken,
			keycloakDataAccess.ClientID,
			keycloakDataAccess.ClientSecret,
			keycloakDataAccess.Realm,
		)

		if err != nil {
			appError := errors.MiddlewareError(err.Error())
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		isTokenValid := *rptResult.Active

		if !isTokenValid {
			appError := errors.InvalidTokenError()
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		token, _, err := services.AuthClient.DecodeAccessToken(
			c,
			accessToken,
			keycloakDataAccess.Realm,
		)

		if err != nil {
			appError := errors.MiddlewareError(err.Error())
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		claims := token.Claims.(jsonToken.MapClaims)

		jsonData, _ := json.Marshal(claims)

		var userClaim entities.JWTClaim

		err = json.Unmarshal(jsonData, &userClaim)
		if err != nil {
			appError := errors.MiddlewareError(err.Error())
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		keyCloakData := config.EnvKeyCloak()

		client := userClaim.ResourceAccess[keyCloakData.ClientID].(map[string]interface{})

		rolesBytes, err := json.Marshal(client["roles"])

		err = json.Unmarshal(rolesBytes, &userClaim.Roles)
		if err != nil {
			appError := errors.MiddlewareError(err.Error())
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		containsRole := userClaim.Roles.Contains(role)

		if !containsRole {
			appError := errors.UnauthorizedError()
			httpError := appError.ToHttpError()
			logger.Log.Error(appError.Message, appError.ToMap())
			c.AbortWithStatusJSON(httpError.StatusCode, httpError)
			c.Abort()
			return
		}

		c.Set("claims", userClaim)
		handler(c)
	}
}
