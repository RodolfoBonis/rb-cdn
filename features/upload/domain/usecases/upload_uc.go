package usecases

import (
	"fmt"
	rbauth "github.com/RodolfoBonis/rb_auth_client"
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
// @Param bucket formData string true "Bucket name"
// @Param folder formData string false "Folder name (optional)"
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} entities.UploadResponseEntity
// @Failure 400 {object} errors.HttpError
// @Failure 401 {object} errors.HttpError
// @Failure 403 {object} errors.HttpError
// @Failure 500 {object} errors.HttpError
// @Router /upload [post]
func (uc *UploadHandler) Upload(c *gin.Context) {
	// Get validation from context
	validation := rbauth.GetValidation(c)
	if validation == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	// Get bucket from form
	bucketName := c.PostForm("bucket")
	if bucketName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bucket parameter is required",
		})
		return
	}

	// Check write permissions for the bucket
	if !validation.Permissions.HasBucketPermission("rb-cdn", bucketName, "write") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("No write permission for bucket: %s", bucketName),
		})
		return
	}

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

	uc.log.Info(fmt.Sprintf("Sending %s to Bucket: %s", objectName, bucketName))
	filePath, appErr := uc.minioService.UploadObject(bucketName, fileEntity, minio.PutObjectOptions{ContentType: contentType})
	if appErr != nil {
		c.JSON(http.StatusInternalServerError, appErr)
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
