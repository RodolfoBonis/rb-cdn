package middlewares

import (
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type MonitoringMiddleware struct {
	logger logger.CustomLogger
}

func NewMonitoringMiddleware(logger logger.CustomLogger) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		logger: logger,
	}
}

func (m *MonitoringMiddleware) SentryMiddleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{Repanic: true})
}

func (m *MonitoringMiddleware) LogMiddleware(ctx *gin.Context) {
	var responseBody = logger.HandleResponseBody(ctx.Writer)
	var requestBody = logger.HandleRequestBody(ctx.Request)
	requestId := uuid.NewString()

	if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetTag("requestId", requestId)
		ctx.Writer = responseBody
	}

	ctx.Next()

	logMessage := logger.FormatRequestAndResponse(ctx.Writer, ctx.Request, responseBody.Body.String(), requestId, requestBody)

	if logMessage != "" {
		if isSuccessStatusCode(ctx.Writer.Status()) {
			m.logger.Info(logMessage)
		} else {
			m.logger.Error(logMessage)
		}
	}
}

func isSuccessStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent, http.StatusPartialContent:
		return true
	default:
		return false
	}
}
