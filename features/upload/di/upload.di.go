package di

import (
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/RodolfoBonis/rb-cdn/features/upload/usecases"
)

func UploadInjection() *usecases.UploadHandler {
	minioService := services.NewMinioService()

	return usecases.NewUploadHandler(minioService, logger.Log)
}
