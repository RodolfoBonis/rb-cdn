package usecases

import (
	"github.com/gin-gonic/gin"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/features/auth/domain/entities"
	"net/http"
	"strings"
)

// RefreshAuthToken godoc
// @Summary Refresh Login Access Token
// @Schemes
// @Description Refresh User Token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} entities.LoginResponseEntity
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 409 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /auth/refresh [post]
func (uc *AuthUseCase) RefreshAuthToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if len(authHeader) < 1 {
		err := errors.InvalidTokenError()
		httpError := err.ToHttpError()
		logger.Log.Error(err.Message, err.ToMap())
		c.AbortWithStatusJSON(httpError.StatusCode, httpError)
		c.Abort()
		return
	}

	refreshToken := strings.Split(authHeader, " ")[1]

	token, err := uc.KeycloakClient.RefreshToken(
		c,
		refreshToken,
		uc.KeycloakAccessData.ClientID,
		uc.KeycloakAccessData.ClientSecret,
		uc.KeycloakAccessData.Realm,
	)

	if err != nil {
		currentError := errors.UsecaseError(err.Error())
		httpError := currentError.ToHttpError()
		logger.Log.Error(currentError.Message, currentError.ToMap())
		c.AbortWithStatusJSON(httpError.StatusCode, httpError)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, entities.LoginResponseEntity{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
	})
}
