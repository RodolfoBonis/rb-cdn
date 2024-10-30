package entities

// LoginResponseEntity model info
// @description LoginResponseEntity model data
type LoginResponseEntity struct {
	// Token to access this API
	AccessToken string `json:"accessToken"`
	// Token to refresh Access Token
	RefreshToken string `json:"refreshToken"`
	// Time to expires token in int
	ExpiresIn int `json:"expiresIn"`
}
