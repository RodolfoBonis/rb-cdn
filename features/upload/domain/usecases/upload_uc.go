package usecases

import (
	"fmt"
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	coreEntities "github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/RodolfoBonis/rb-cdn/features/upload/domain/entities"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"net/http"
	"strings"
)

type UploadHandler struct {
	minioService services.IMinioService
	log          *logger.CustomLogger
}

func NewUploadHandler(minioService services.IMinioService, log *logger.CustomLogger) *UploadHandler {
	return &UploadHandler{minioService: minioService, log: log}
}

// Upload godoc
// @Summary Upload a file to CDN
// @Description Uploads a file to the CDN storage and returns the access URL
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param folder formData string false "Folder name (optional)"
// @Param X-API-KEY header string true "API Key for authentication"
// @Success 200 {object} entities.UploadResponseEntity
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /upload [post]
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

	fileEntity := coreEntities.FileEntity{
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

	uc.log.Info(fmt.Sprintf("Sending %s to Bucket: %s", objectName, apiKeyData.Bucket))
	filePath, appErr := uc.minioService.UploadObject(apiKeyData.Bucket, fileEntity, minio.PutObjectOptions{ContentType: contentType})
	if appErr != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Erro ao fazer upload: %s", appErr))
		return
	}

	extension := strings.Split(objectName, ".")[1]

	videoExtensions := map[string]bool{
		"mp4": true,
		"avi": true,
		"mkv": true,
		"mov": true,
		"flv": true,
		"wmv": true,
	}

	message := fmt.Sprintf("Arquivo '%s' enviado com sucesso!", objectName)
	rootUri := "https://rb-cdn.rodolfodebonis.com.br/v1"
	if videoExtensions[extension] {
		c.JSON(http.StatusOK, entities.UploadResponseEntity{
			URL:     fmt.Sprintf("%s/stream/%s", rootUri, objectName),
			Message: message,
		})
		return
	}

	c.JSON(http.StatusOK, entities.UploadResponseEntity{
		URL:     fmt.Sprintf("%s/cdn/%s", rootUri, filePath),
		Message: message,
	})
}
