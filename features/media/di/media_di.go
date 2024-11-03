package di

import (
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/RodolfoBonis/rb-cdn/features/media/usecases"
)

func MediaInjection() *usecases.MediaHandler {
	minioService := services.NewMinioService()
	return usecases.NewMediaHandler(minioService)
}
