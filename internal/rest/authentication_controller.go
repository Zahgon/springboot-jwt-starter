package rest

import (
	"encoding/json"
	"net/http"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/security"
	"github.com/bfwg/springboot-jwt-starter/internal/security/auth"
	"github.com/bfwg/springboot-jwt-starter/internal/service"
)

// AuthenticationController serves the /auth endpoints.
type AuthenticationController struct {
	tokens *security.TokenHelper
	users  service.UserService
}

// NewAuthenticationController builds the controller.
func NewAuthenticationController(tokens *security.TokenHelper, users service.UserService) *AuthenticationController {
	return &AuthenticationController{tokens: tokens, users: users}
}

// CreateAuthenticationToken handles POST /auth/login. A malformed body, an
// unknown username and a wrong password are all indistinguishable: each is
// turned away through the shared 401 entry point.
func (c *AuthenticationController) CreateAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var request auth.JwtAuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		auth.Commence(w)
		return
	}

	user, err := c.users.Authenticate(request.Username, request.Password)
	if err != nil {
		auth.Commence(w)
		return
	}

	jws := c.tokens.GenerateToken(user.GetUsername())
	expiresIn := c.tokens.GetExpiredIn()
	writeJSON(w, http.StatusOK, model.NewUserTokenState(jws, int64(expiresIn)))
}

// RefreshAuthenticationToken handles POST /auth/refresh.
//
// With no authenticated principal the answer is 202 and an all-null token
// state. With a principal the answer is 200 and the configured expiry, even
// when the incoming token could not be refreshed and the token field is null.
func (c *AuthenticationController) RefreshAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	authToken := c.tokens.GetToken(r)
	principal := auth.FromRequest(r)

	if authToken == "" || principal == nil {
		writeJSON(w, http.StatusAccepted, model.EmptyUserTokenState())
		return
	}

	refreshed := c.tokens.RefreshToken(authToken)
	expiresIn := c.tokens.GetExpiredIn()
	writeJSON(w, http.StatusOK, model.NewUserTokenStateWithExpiry(refreshed, int64(expiresIn)))
}

// passwordChanger is the change-password request body.
type passwordChanger struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangePassword handles POST /auth/change-password for a caller holding
// ROLE_USER. A wrong old password is a credentials failure and answers 401.
func (c *AuthenticationController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	principal := auth.FromRequest(r)
	if principal == nil {
		auth.Commence(w)
		return
	}

	var request passwordChanger
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		auth.Commence(w)
		return
	}

	if err := c.users.ChangePassword(principal.Name(), request.OldPassword, request.NewPassword); err != nil {
		auth.Commence(w)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"result": "success"})
}
