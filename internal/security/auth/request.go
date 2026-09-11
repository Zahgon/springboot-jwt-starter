package auth

// JwtAuthenticationRequest is the login request body.
type JwtAuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
