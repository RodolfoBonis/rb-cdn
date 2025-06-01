package errors

import (
	"testing"

	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	t.Run("DatabaseError", func(t *testing.T) {
		err := DatabaseError("database error")
		assert.Equal(t, entities.AppError.Database, err.Error)
		assert.Equal(t, "database error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("EntityError", func(t *testing.T) {
		err := EntityError("entity error")
		assert.Equal(t, entities.AppError.Entity, err.Error)
		assert.Equal(t, "entity error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("EnvironmentError", func(t *testing.T) {
		err := EnvironmentError("environment error")
		assert.Equal(t, entities.AppError.Environment, err.Error)
		assert.Equal(t, "environment error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("MiddlewareError", func(t *testing.T) {
		err := MiddlewareError("middleware error")
		assert.Equal(t, entities.AppError.Middleware, err.Error)
		assert.Equal(t, "middleware error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("ModelError", func(t *testing.T) {
		err := ModelError("model error")
		assert.Equal(t, entities.AppError.Model, err.Error)
		assert.Equal(t, "model error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		err := RepositoryError("repository error")
		assert.Equal(t, entities.AppError.Repository, err.Error)
		assert.Equal(t, "repository error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("RootError", func(t *testing.T) {
		err := RootError("root error")
		assert.Equal(t, entities.AppError.Root, err.Error)
		assert.Equal(t, "root error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("ServiceError", func(t *testing.T) {
		err := ServiceError("service error")
		assert.Equal(t, entities.AppError.Service, err.Error)
		assert.Equal(t, "service error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		err := UsecaseError("usecase error")
		assert.Equal(t, entities.AppError.Usecase, err.Error)
		assert.Equal(t, "usecase error", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("NotFoundError", func(t *testing.T) {
		err := NotFoundError()
		assert.Equal(t, entities.AppError.NotFound, err.Error)
		assert.Equal(t, "Resource not found", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("InvalidTokenError", func(t *testing.T) {
		err := InvalidTokenError()
		assert.Equal(t, entities.AppError.InvalidToken, err.Error)
		assert.Equal(t, "Invalid Token", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("InvalidCredentialsError", func(t *testing.T) {
		err := InvalidCredentialsError()
		assert.Equal(t, entities.AppError.InvalidCredentials, err.Error)
		assert.Equal(t, "Invalid Credentials", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("UnauthorizedError", func(t *testing.T) {
		err := UnauthorizedError()
		assert.Equal(t, entities.AppError.Unauthorized, err.Error)
		assert.Equal(t, "Unauthorized", err.Message)
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("ToMap", func(t *testing.T) {
		err := DatabaseError("test error")
		errMap := err.ToMap()

		assert.Equal(t, err.Error, errMap["error"])
		assert.Equal(t, err.Message, errMap["message"])
		assert.NotEmpty(t, errMap["stack_trace"])
	})

	t.Run("ToHttpError", func(t *testing.T) {
		err := DatabaseError("test error")
		httpErr := err.ToHttpError()

		assert.Equal(t, entities.AppErrorToHTTPCode[err.Error], httpErr.StatusCode)
		assert.Equal(t, err.Message, httpErr.Message)
	})
}
