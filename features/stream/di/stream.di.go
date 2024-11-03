package di

import (
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/RodolfoBonis/rb-cdn/features/stream/domain/usecases"
)

func StreamInjection() *usecases.StreamHandler {
	minioService := services.NewMinioService()
	return usecases.NewStreamHandler(minioService, logger.Log)
}
