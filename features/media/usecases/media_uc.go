package usecases

import (
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

func (uc *MediaHandler) Media(c *gin.Context) {
	objectName := c.Param("objectPath")[1:]
	bucket := c.Param("bucket")

	if objectName == "" || bucket == "" {
		c.String(http.StatusBadRequest, "Invalid object path")
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
