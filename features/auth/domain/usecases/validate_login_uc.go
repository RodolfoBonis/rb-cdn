package usecases

import (
	"github.com/gin-gonic/gin"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/features/auth/domain/entities"
	"net/http"
	"strings"
)

// ValidateLogin godoc
// @Summary Validate auth
// @Schemes
// @Description performs auth of user
// @Tags Auth
// @Accept json
// @Produce json
// @Param _ body entities.RequestLoginEntity true "Login Data"
// @Success 200 {object} entities.LoginResponseEntity
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 409 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /auth [post]
func (uc *AuthUseCase) ValidateLogin(c *gin.Context) {
	loginData := new(entities.RequestLoginEntity)
	logger.Log.Info("Validating login", map[string]interface{}{
		"email": loginData.Email,
	})
	err := c.BindJSON(&loginData)

	if err != nil {
		internalError := errors.UsecaseError(err.Error())
		httpError := internalError.ToHttpError()
		logger.Log.Error(internalError.Message, internalError.ToMap())
		c.AbortWithStatusJSON(httpError.StatusCode, httpError)
		return
	}

	jwt, err := uc.KeycloakClient.Login(
		c,
		uc.KeycloakAccessData.ClientID,
		uc.KeycloakAccessData.ClientSecret,
		uc.KeycloakAccessData.Realm,
		loginData.Email,
		loginData.Password,
	)

	if err != nil {
		internalError := errors.InvalidCredentialsError()
		httpError := internalError.ToHttpError()
		logger.Log.Error(internalError.Message, internalError.ToMap())
		c.AbortWithStatusJSON(httpError.StatusCode, httpError)
		return
	}

	logger.Log.Info("Login successful", map[string]interface{}{
		"email": loginData.Email,
	})

	c.JSON(http.StatusOK, entities.LoginResponseEntity{
		AccessToken:  jwt.AccessToken,
		RefreshToken: jwt.RefreshToken,
		ExpiresIn:    jwt.ExpiresIn,
	})
}

// ValidateToken godoc
// @Summary Validate Auth Token
// @Schemes
// @Description Validate current Auth Token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} bool
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 409 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /auth/validate [post]
func (uc *AuthUseCase) ValidateToken(c *gin.Context) {
	authorization := c.GetHeader("Authorization")
	logger.Log.Info("Validating token")

	token := strings.Split(authorization, " ")[1]

	rptResult, err := uc.KeycloakClient.RetrospectToken(
		c,
		token,
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

	logger.Log.Info("Token validated")

	isTokenValid := *rptResult.Active

	if !isTokenValid {
		currentError := errors.InvalidTokenError()
		httpError := currentError.ToHttpError()
		logger.Log.Error(currentError.Message, currentError.ToMap())
		c.AbortWithStatusJSON(httpError.StatusCode, httpError)
		c.Abort()
		return
	}

	logger.Log.Info("Token is valid")

	c.JSON(http.StatusOK, true)
}
