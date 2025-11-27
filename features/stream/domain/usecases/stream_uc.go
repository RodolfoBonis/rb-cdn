package usecases

import (
	"fmt"
	rbauth "github.com/RodolfoBonis/rb_auth_client"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
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

// StreamVideo godoc
// @Summary Stream video content
// @Schemes
// @Description Streams video content from MinIO with support for range requests
// @Tags Stream
// @Accept json
// @Produce video/mp4
// @Param objectPath path string true "Object path in the bucket"
// @Param Range header string false "Range header for partial content requests"
// @Param Authorization header string true "Bearer token"
// @Success 200 {file} binary "Full video content"
// @Success 206 {file} binary "Partial video content"
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /stream/{objectPath} [get]
func (vc *StreamHandler) StreamVideo(c *gin.Context) {
	// Get validation from context
	validation := rbauth.GetValidation(c)
	if validation == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	objectName := c.Param("objectPath")[1:]

	// Extract bucket from object path (assuming format: bucket/path/to/file)
	// For stream, we need to determine which bucket to use
	// We'll check all buckets the user has read access to
	var bucketName string
	for service, buckets := range validation.Permissions {
		if service == "rb-cdn" {
			for bucket, permissions := range buckets {
				for _, perm := range permissions {
					if perm == "read" {
						bucketName = bucket
						break
					}
				}
				if bucketName != "" {
					break
				}
			}
		}
		if bucketName != "" {
			break
		}
	}

	if bucketName == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "No read permission for any bucket",
		})
		return
	}

	// Check read permissions
	if !validation.Permissions.HasBucketPermission("rb-cdn", bucketName, "read") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("No read permission for bucket: %s", bucketName),
		})
		return
	}

	obj, appErr := vc.getMinioObject(c, bucketName, objectName)
	if appErr != nil {
		return
	}
	defer obj.Close()

	objInfo, err := vc.getObjectInfo(c, obj)
	if err != nil {
		return
	}

	contentLength := objInfo.Size
	vc.setCommonHeaders(c, contentLength)

	rangeHeader := c.Request.Header.Get("Range")
	if rangeHeader == "" {
		vc.handleFullContent(c, obj, contentLength)
		return
	}

	vc.handleRangeRequest(c, obj, rangeHeader, contentLength)
}

func (vc *StreamHandler) getMinioObject(c *gin.Context, bucket, objectName string) (*minio.Object, *errors.AppError) {
	obj, appErr := vc.minioService.GetObject(bucket, objectName, minio.GetObjectOptions{})
	if appErr != nil {
		vc.logger.Error(fmt.Sprintf("Erro ao obter o objeto do MinIO: %v", appErr))
		c.AbortWithStatusJSON(http.StatusInternalServerError, appErr)
		return nil, appErr
	}
	return obj, nil
}

func (vc *StreamHandler) getObjectInfo(c *gin.Context, obj *minio.Object) (minio.ObjectInfo, error) {
	objInfo, err := obj.Stat()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter o tamanho do vídeo"})
		return minio.ObjectInfo{}, err
	}
	return objInfo, nil
}

func (vc *StreamHandler) setCommonHeaders(c *gin.Context, contentLength int64) {
	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
}

func (vc *StreamHandler) handleFullContent(c *gin.Context, obj *minio.Object, contentLength int64) {
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, obj); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
	}
}

func (vc *StreamHandler) handleRangeRequest(c *gin.Context, obj *minio.Object, rangeHeader string, contentLength int64) {
	start, end := vc.parseRange(rangeHeader, contentLength)
	partSize := end - start + 1

	vc.setRangeHeaders(c, start, end, contentLength, partSize)

	if err := vc.streamRange(c, obj, start, partSize); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
	}
}

func (vc *StreamHandler) parseRange(rangeHeader string, contentLength int64) (start, end int64) {
	_, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if err != nil {
		return 0, 0
	}
	if end == 0 || end >= contentLength {
		end = contentLength - 1
	}
	return
}

func (vc *StreamHandler) setRangeHeaders(c *gin.Context, start, end, contentLength, partSize int64) {
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, contentLength))
	c.Header("Content-Length", strconv.FormatInt(partSize, 10))
	c.Status(http.StatusPartialContent)
}

func (vc *StreamHandler) streamRange(c *gin.Context, obj *minio.Object, start, partSize int64) error {
	if _, err := obj.Seek(start, 0); err != nil {
		return err
	}

	buf := make([]byte, 4096)
	bytesSent := int64(0)

	for bytesSent < partSize {
		bytesToRead := vc.calculateBytesToRead(buf, bytesSent, partSize)
		if err := vc.writeChunk(c, obj, buf, bytesToRead, &bytesSent); err != nil {
			return err
		}
	}
	return nil
}

func (vc *StreamHandler) calculateBytesToRead(buf []byte, bytesSent, partSize int64) int64 {
	bytesToRead := int64(len(buf))
	if bytesToRead > partSize-bytesSent {
		bytesToRead = partSize - bytesSent
	}
	return bytesToRead
}

func (vc *StreamHandler) writeChunk(c *gin.Context, obj *minio.Object, buf []byte, bytesToRead int64, bytesSent *int64) error {
	n, err := obj.Read(buf[:bytesToRead])
	if err != nil {
		return err
	}

	if _, err := c.Writer.Write(buf[:n]); err != nil {
		return err
	}

	*bytesSent += int64(n)
	return nil
}
