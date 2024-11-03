package usecases

import (
	"fmt"
	keyGuardian "github.com/RodolfoBonis/go_key_guardian"
	"github.com/RodolfoBonis/rb-cdn/core/logger"
	"github.com/RodolfoBonis/rb-cdn/core/services"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go"
	"io"
	"net/http"
	"strconv"
)

type StreamHandler struct {
	minioService services.IMinioService
	logger       *logger.CustomLogger
}

func NewStreamHandler(minioService services.IMinioService, logger *logger.CustomLogger) *StreamHandler {
	return &StreamHandler{minioService: minioService, logger: logger}
}

func (vc *StreamHandler) StreamVideo(c *gin.Context) {
	objectName := c.Param("objectPath")[1:]

	data, exists := c.Get("configs")
	if !exists {
		c.String(http.StatusInternalServerError, "Erro ao obter as configurações")
		return
	}

	apiKeyData := data.(keyGuardian.ApiKeyData)

	obj, appErr := vc.minioService.GetObject(apiKeyData.Bucket, objectName, minio.GetObjectOptions{})

	if appErr != nil {
		vc.logger.Error(fmt.Sprintf("Erro ao obter o objeto do MinIO: %v", appErr))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erro ao acessar o vídeo"})
		return
	}
	defer func(obj *minio.Object) {
		err := obj.Close()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		}
	}(obj)

	objInfo, err := obj.Stat()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter o tamanho do vídeo"})
		return
	}

	contentLength := objInfo.Size

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.Request.Header.Get("Range")
	if rangeHeader == "" {
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		c.Status(http.StatusOK)
		_, err = io.Copy(c.Writer, obj)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		return
	}

	var start, end int64
	_, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if err != nil || end == 0 {
		end = contentLength - 1
	}
	if end >= contentLength {
		end = contentLength - 1
	}
	partSize := end - start + 1

	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, contentLength))
	c.Header("Content-Length", strconv.FormatInt(partSize, 10))
	c.Status(http.StatusPartialContent)

	_, err = obj.Seek(start, 0)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	buf := make([]byte, 4096)
	bytesSent := int64(0)
	for bytesSent < partSize {
		bytesToRead := int64(len(buf))
		if bytesToRead > partSize-bytesSent {
			bytesToRead = partSize - bytesSent
		}
		n, err := obj.Read(buf[:bytesToRead])
		if err != nil {
			break
		}
		_, err = c.Writer.Write(buf[:n])
		if err != nil {
			break
		}
		bytesSent += int64(n)

		if bytesSent >= partSize {
			break
		}
	}
}
