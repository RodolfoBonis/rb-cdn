package entities

import "net/http"

type appErrorTypes struct {
	Database           int
	Repository         int
	Usecase            int
	Entity             int
	Model              int
	Service            int
	Middleware         int
	Root               int
	Environment        int
	NotFound           int
	InvalidToken       int
	InvalidCredentials int
	Unauthorized       int
}

var AppError = appErrorTypes{
	Database:           1001,
	Repository:         1002,
	Usecase:            1003,
	Entity:             1004,
	Model:              1005,
	Service:            1006,
	Middleware:         1007,
	Root:               1008,
	Environment:        1009,
	NotFound:           1010,
	InvalidToken:       1011,
	InvalidCredentials: 1012,
	Unauthorized:       1013,
}

var AppErrorToHTTPCode = map[int]int{
	AppError.Database:           http.StatusInternalServerError, // Database
	AppError.Repository:         http.StatusInternalServerError, // Repository
	AppError.Usecase:            http.StatusInternalServerError, // Usecase
	AppError.Entity:             http.StatusBadRequest,          // Entity
	AppError.Model:              http.StatusBadRequest,          // Model
	AppError.Service:            http.StatusInternalServerError, // Service
	AppError.Middleware:         http.StatusInternalServerError, // Middleware
	AppError.Root:               http.StatusInternalServerError, // Root
	AppError.Environment:        http.StatusInternalServerError, // Environment
	AppError.NotFound:           http.StatusNotFound,            // NotFound
	AppError.InvalidToken:       http.StatusUnauthorized,        // InvalidToken
	AppError.InvalidCredentials: http.StatusUnauthorized,        // InvalidCredentials
	AppError.Unauthorized:       http.StatusUnauthorized,        // Unauthorized
}
