package errors

import (
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/types"
)

type AppError struct {
	Error      int              `json:"app_error"`
	Message    string           `json:"message"`
	StackTrace types.StackTrace `json:"stack_trace"`
}

func (e *AppError) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"error":       e.Error,
		"message":     e.Message,
		"stack_trace": e.StackTrace.String(),
	}
}

func (e *AppError) ToHttpError() *HttpError {
	return NewHTTPError(entities.AppErrorToHTTPCode[e.Error], e.Message)
}

func newAppError(error int, message string) *AppError {
	return &AppError{
		Error:      error,
		Message:    message,
		StackTrace: callers(),
	}
}

func DatabaseError(message string) *AppError {
	return newAppError(
		entities.AppError.Database,
		message,
	)
}

func EntityError(message string) *AppError {
	return newAppError(
		entities.AppError.Entity,
		message,
	)
}

func EnvironmentError(message string) *AppError {
	return newAppError(
		entities.AppError.Environment,
		message,
	)
}

func MiddlewareError(message string) *AppError {
	return newAppError(
		entities.AppError.Middleware,
		message,
	)
}

func ModelError(message string) *AppError {
	return newAppError(
		entities.AppError.Model,
		message,
	)
}

func RepositoryError(message string) *AppError {
	return newAppError(
		entities.AppError.Repository,
		message,
	)
}

func RootError(message string) *AppError {
	return newAppError(
		entities.AppError.Root,
		message,
	)
}

func ServiceError(message string) *AppError {
	return newAppError(
		entities.AppError.Service,
		message,
	)
}

func UsecaseError(message string) *AppError {
	return newAppError(
		entities.AppError.Usecase,
		message,
	)
}

func NotFoundError() *AppError {
	return newAppError(
		entities.AppError.NotFound,
		"Resource not found",
	)
}

func InvalidTokenError() *AppError {
	return newAppError(
		entities.AppError.InvalidToken,
		"Invalid Token",
	)
}

func InvalidCredentialsError() *AppError {
	return newAppError(
		entities.AppError.InvalidCredentials,
		"Invalid Credentials",
	)
}

func UnauthorizedError() *AppError {
	return newAppError(
		entities.AppError.Unauthorized,
		"Unauthorized",
	)
}
