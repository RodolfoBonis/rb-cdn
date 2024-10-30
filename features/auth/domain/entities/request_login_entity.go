package entities

// RequestLoginEntity model info
// @description RequestLoginEntity model data
type RequestLoginEntity struct {
	// User email
	Email string `json:"email"`
	// User password
	Password string `json:"password"`
}
