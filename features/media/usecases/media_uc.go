package usecases

import (
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"io"
	"net/http"
	"strings"
)

type MediaHandler struct {
	minioService services.IMinioService
}

func NewMediaHandler(minioService services.IMinioService) *MediaHandler {
	return &MediaHandler{minioService: minioService}
}

// Media godoc
// @Summary Get media from CDN
// @Description Retrieves media files from the CDN, supporting images and videos
// @Tags Media
// @Accept json
// @Produce octet-stream
// @Produce image/jpeg
// @Produce image/png
// @Param bucket path string true "Bucket name"
// @Param objectPath path string true "Path to the object in the bucket"
// @Param X-API-KEY header string true "API Key for authentication"
// @Success 200 {file} binary "Media file"
// @Success 307 {object} errors.HttpError
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 204 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /cdn/{bucket}/{objectPath} [get]
func (uc *MediaHandler) Media(c *gin.Context) {
	objectName := c.Param("objectPath")[1:]
	bucket := c.Param("bucket")

	if objectName == "" || bucket == "" {
		c.String(http.StatusBadRequest, "Invalid object path")
		return
	}

	data, exists := c.Get("configs")
	if !exists {
		c.String(http.StatusInternalServerError, "Erro ao obter as configurações")
		return
	}

	apiKeyData := data.(keyGuardian.ApiKeyData)

	if apiKeyData.Bucket != bucket {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	object, appError := uc.minioService.GetObject(bucket, objectName, minio.GetObjectOptions{})

	if appError != nil {
		c.String(http.StatusNoContent, "Error while getting object")
		return
	}

	defer func(object *minio.Object) {
		err := object.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error while closing object")
		}
	}(object)

	extension := strings.Split(objectName, ".")[1]

	videoExtensions := map[string]bool{
		"mp4": true,
		"avi": true,
		"mkv": true,
		"mov": true,
		"flv": true,
		"wmv": true,
	}

	if videoExtensions[extension] {
		c.Redirect(http.StatusTemporaryRedirect, "/stream/"+objectName)
		return
	}

	contentType := "application/octet-stream"

	extensionToContentType := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
	}

	if ct, found := extensionToContentType[extension]; found {
		contentType = ct
	}

	c.Header("Content-Type", contentType)

	_, err := io.Copy(c.Writer, object)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	return
}
