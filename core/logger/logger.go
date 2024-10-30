package logger

import (
	_ "fmt"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"go.uber.org/zap"
	"time"
)

var (
	Log *CustomLogger
)

// CustomLogger é uma estrutura que encapsula um zap.Logger.
type CustomLogger struct {
	logger *zap.Logger
}

// LogData encapsula os dados do log.
type LogData struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Time    time.Time              `json:"time"`
	JSON    map[string]interface{} `json:"json,omitempty"`
}

// InitLogger cria uma nova instância do CustomLogger.
func InitLogger() {
	if config.EnvironmentConfig() == entities.Environment.Development {
		Log = &CustomLogger{
			logger: config.ZapTestConfig(),
		}

		return
	}

	newRelicConfig := config.NewRelicConfig()

	Log = &CustomLogger{
		logger: config.ZapConfig(newRelicConfig),
	}
}

// Info envia um log de informação para o New Relic e o logger.
func (cl *CustomLogger) Info(message string, jsonData ...map[string]interface{}) {
	cl.logger.Info(message, zap.Any("json", jsonData))
}

// Warning envia um log de aviso para o New Relic e o logger.
func (cl *CustomLogger) Warning(message string, jsonData ...map[string]interface{}) {
	cl.logger.Warn(message, zap.Any("json", jsonData))
}

// Error envia um log de erro para o New Relic e o logger.
func (cl *CustomLogger) Error(message string, jsonData ...map[string]interface{}) {
	cl.logger.Error(message, zap.Any("json", jsonData))
}
