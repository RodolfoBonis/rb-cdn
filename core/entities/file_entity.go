package entities

import "mime/multipart"

type FileEntity struct {
	File multipart.File `json:"file,omitempty" validate:"required"`
	Size int64          `json:"size" validate:"required"`
	Name string         `json:"name" validate:"required"`
}
