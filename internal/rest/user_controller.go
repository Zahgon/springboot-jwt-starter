package rest

import (
	"net/http"
	"strconv"

	"github.com/bfwg/springboot-jwt-starter/internal/security/auth"
	"github.com/bfwg/springboot-jwt-starter/internal/service"
)

// UserController serves the /api endpoints.
type UserController struct {
	users service.UserService
}

// NewUserController builds the controller.
func NewUserController(users service.UserService) *UserController {
	return &UserController{users: users}
}

// LoadByID handles GET /api/user/{userId} for a caller holding ROLE_ADMIN.
// A userId that is not a number is turned away as 401, and a userId with no
// account answers 200 with an empty body.
func (c *UserController) LoadByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		auth.Commence(w)
		return
	}

	user, err := c.users.FindByID(id)
	if err != nil {
		auth.Commence(w)
		return
	}
	if user == nil {
		writeEmptyOK(w)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// LoadAll handles GET /api/user/all for a caller holding ROLE_ADMIN.
func (c *UserController) LoadAll(w http.ResponseWriter, r *http.Request) {
	users, err := c.users.FindAll()
	if err != nil {
		auth.Commence(w)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// User handles GET /api/whoami for a caller holding ROLE_USER.
//
// The account is looked up by name rather than taken from the principal, which
// keeps the role guard meaningful for this endpoint.
func (c *UserController) User(w http.ResponseWriter, r *http.Request) {
	principal := auth.FromRequest(r)
	if principal == nil {
		auth.Commence(w)
		return
	}

	user, err := c.users.FindByUsername(principal.Name())
	if err != nil {
		auth.Commence(w)
		return
	}
	if user == nil {
		writeEmptyOK(w)
		return
	}
	writeJSON(w, http.StatusOK, user)
}
