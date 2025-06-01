package services

import (
	"fmt"
	"github.com/RodolfoBonis/rb-cdn/core/config"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
	"github.com/RodolfoBonis/rb-cdn/core/errors"
	"github.com/minio/minio-go"
	"net/url"
	"time"
)

type MinioService struct {
	host      string
	accessId  string
	secretKey string
}

type IMinioService interface {
	startMinioService() (*minio.Client, *errors.AppError)
	UploadObject(bucket string, file entities.FileEntity, options minio.PutObjectOptions) (string, *errors.AppError)
	GetObject(bucket string, objectName string, options minio.GetObjectOptions) (*minio.Object, *errors.AppError)
	GetObjectURL(bucket string, objectName string) (string, *errors.AppError)
	GetObjectInfo(bucket string, objectName string) (*minio.ObjectInfo, *errors.AppError)
}

func NewMinioService() IMinioService {
	return &MinioService{
		host:      config.EnvMinioHost(),
		accessId:  config.EnvMinioAccessId(),
		secretKey: config.EnvMinioSecretKey(),
	}
}

func (service *MinioService) startMinioService() (*minio.Client, *errors.AppError) {

	client, err := minio.New(service.host, service.accessId, service.secretKey, true)

	if err != nil {
		return nil, errors.ServiceError(err.Error())
	}

	return client, nil
}

func (service *MinioService) checkIfBucketExists(bucket string) bool {
	client, appErr := service.startMinioService()

	if appErr != nil {
		panic(appErr)
	}

	exists, err := client.BucketExists(bucket)

	if err != nil {
		return false
	}

	return exists
}

func (service *MinioService) UploadObject(bucket string, file entities.FileEntity, options minio.PutObjectOptions) (string, *errors.AppError) {

	bucketExists := service.checkIfBucketExists(bucket)

	if !bucketExists {
		return "", errors.ServiceError("Bucket does not exist")
	}

	client, appError := service.startMinioService()

	if appError != nil {
		return "", appError
	}

	_, err := client.PutObject(
		bucket,
		file.Name,
		file.File,
		file.Size,
		options,
	)

	if err != nil {
		return "", errors.ServiceError(err.Error())
	}

	return fmt.Sprintf("%s/%s", bucket, file.Name), nil
}

func (service *MinioService) GetObject(bucket string, objectName string, options minio.GetObjectOptions) (*minio.Object, *errors.AppError) {
	client, appError := service.startMinioService()

	if appError != nil {
		return nil, appError
	}

	object, err := client.GetObject(bucket, objectName, options)
	if err != nil {
		return nil, errors.ServiceError(err.Error())
	}

	return object, nil
}

func (service *MinioService) GetObjectURL(bucket string, objectName string) (string, *errors.AppError) {
	client, appError := service.startMinioService()
	if appError != nil {
		return "", appError
	}

	reqParams := make(url.Values)
	presignedURL, err := client.PresignedGetObject(bucket, objectName, time.Hour, reqParams)
	if err != nil {
		return "", errors.ServiceError(err.Error())
	}

	return presignedURL.String(), nil
}

func (service *MinioService) GetObjectInfo(bucket string, objectName string) (*minio.ObjectInfo, *errors.AppError) {
	client, appError := service.startMinioService()
	if appError != nil {
		return nil, appError
	}

	objectInfo, err := client.StatObject(bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, errors.ServiceError(err.Error())
	}

	return &objectInfo, nil
}
