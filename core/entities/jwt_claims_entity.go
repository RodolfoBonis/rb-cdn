package entities

import (
	"github.com/google/uuid"
	"github.com/RodolfoBonis/rb-cdn/core/types"
)

type JWTClaim struct {
	ID             uuid.UUID              `json:"sub"`
	Verified       bool                   `json:"email_verified"`
	Name           string                 `json:"name"`
	Username       string                 `json:"preferred_username"`
	FirstName      string                 `json:"given_name"`
	FamilyName     string                 `json:"family_name"`
	Email          string                 `json:"email"`
	ResourceAccess map[string]interface{} `json:"resource_access,omitempty"`
	Roles          types.Array            `json:"roles,omitempty"`
}
