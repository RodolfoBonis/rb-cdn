package services

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/RodolfoBonis/rb-cdn/core/config"
)

var AuthClient *gocloak.GoCloak

func InitializeOAuthServer() {
	keycloakDataAccess := config.EnvKeyCloak()

	AuthClient = gocloak.NewClient(keycloakDataAccess.Host)
}
