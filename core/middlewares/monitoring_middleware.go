package middlewares

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"net/http"
)

type MonitoringMiddleware struct {
	newRelicConfig *newrelic.Application
}

func NewMonitoringMiddleware(newRelicConfig *newrelic.Application) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		newRelicConfig: newRelicConfig,
	}
}

func (m *MonitoringMiddleware) SentryMiddleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{Repanic: true})
}

func (m *MonitoringMiddleware) NewRelicMiddleware() gin.HandlerFunc {
	return nrgin.Middleware(m.newRelicConfig)
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
			logger.Log.Info(logMessage)
		} else {
			logger.Log.Error(logMessage)
		}
	}
}

func isSuccessStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return true
	default:
		return false
	}
}
