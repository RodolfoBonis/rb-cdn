package usecases

import (
	"github.com/gin-gonic/gin"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"net/http"
	"strings"
)

// Logout godoc
// @Summary Logout
// @Schemes
// @Description Logout the User
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} bool
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 409 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /auth/logout [post]
func (uc *AuthUseCase) Logout(c *gin.Context) {
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

	err := uc.KeycloakClient.Logout(
		c,
		uc.KeycloakAccessData.ClientID,
		uc.KeycloakAccessData.ClientSecret,
		uc.KeycloakAccessData.Realm,
		refreshToken,
	)

	if err != nil {
		currentError := errors.UsecaseError(err.Error())
		httpError := currentError.ToHttpError()
		logger.Log.Error(currentError.Message, currentError.ToMap())
		c.AbortWithStatusJSON(httpError.StatusCode, httpError)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, true)
}
