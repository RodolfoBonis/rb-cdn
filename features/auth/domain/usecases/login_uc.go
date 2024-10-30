package usecases

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/RodolfoBonis/rb-cdn/core/entities"
)

type AuthUseCase struct {
	KeycloakClient     *gocloak.GoCloak
	KeycloakAccessData entities.KeyCloakDataEntity
}
