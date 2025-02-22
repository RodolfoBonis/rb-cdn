package errors

import (
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
)

// swaggo:generate
type HttpError struct {
	StatusCode int    `json:"code"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace,omitempty"`
}

var getEnvironment = config.EnvironmentConfig

func (e *HttpError) ToMap() map[string]interface{} {
	stackTrace := callers()
	return map[string]interface{}{
		"code":        e.StatusCode,
		"message":     e.Message,
		"stack_trace": stackTrace.String(),
	}
}

func NewHTTPError(statusCode int, message string) *HttpError {
	httpError := &HttpError{
		StatusCode: statusCode,
		Message:    message,
	}

	if getEnvironment() == entities.Environment.Development {
		stacktrace := callers()
		httpError.StackTrace = stacktrace.String()
	}

	return httpError
}
