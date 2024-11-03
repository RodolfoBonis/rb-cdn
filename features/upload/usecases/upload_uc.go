package usecases

import (
	"fmt"
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"net/http"
)

type UploadHandler struct {
	minioService services.IMinioService
}

func NewUploadHandler(minioService services.IMinioService) *UploadHandler {
	return &UploadHandler{minioService: minioService}
}

func (uc *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	folderName := c.Request.FormValue("folder")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Erro ao obter o arquivo: %s", err))
		return
	}
	defer file.Close()

	reader, err := header.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Erro ao abrir o arquivo: %s", err))
		return
	}
	defer reader.Close()

	objectName := header.Filename
	contentType := header.Header.Get("Content-Type")
	fileSize := header.Size

	fileNameLocation := fmt.Sprintf("%s/%s", folderName, objectName)

	if folderName == "" {
		fileNameLocation = objectName
	}

	fileEntity := entities.FileEntity{
		File: file,
		Name: fileNameLocation,
		Size: fileSize,
	}

	data, exists := c.Get("configs")
	if !exists {
		c.String(http.StatusInternalServerError, "Erro ao obter as configurações")
		return
	}

	apiKeyData := data.(keyGuardian.ApiKeyData)

	uri, appErr := uc.minioService.UploadObject(apiKeyData.Bucket, fileEntity, minio.PutObjectOptions{ContentType: contentType})
	if appErr != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Erro ao fazer upload: %s", appErr))
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("Arquivo '%s' enviado com sucesso! URL = %s", objectName, uri))
}
