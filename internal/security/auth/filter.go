package auth

import (
	"net/http"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

// TokenReader is the part of the token helper this filter needs.
type TokenReader interface {
	GetToken(r *http.Request) string
	GetUsernameFromToken(token string) string
	ValidateToken(token string, userDetails model.UserDetails) bool
}

// UserDetailsLoader resolves a username to a principal. A username with no
// account yields a nil principal.
type UserDetailsLoader interface {
	LoadUserByUsername(username string) (model.UserDetails, error)
}

// TokenAuthenticationFilter establishes the principal for a request that
// carries a valid bearer token. A request without a token, or with one that
// does not validate, simply continues unauthenticated — it is the route guards,
// not this filter, that reject it.
func TokenAuthenticationFilter(tokens TokenReader, users UserDetailsLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := tokens.GetToken(r); token != "" {
				if username := tokens.GetUsernameFromToken(token); username != "" {
					userDetails, err := users.LoadUserByUsername(username)
					if err == nil && userDetails != nil && tokens.ValidateToken(token, userDetails) {
						r = WithAuthentication(r, &TokenBasedAuthentication{
							Principal: userDetails,
							Token:     token,
						})
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole guards a handler, admitting only a principal holding the role.
// Anything else — anonymous, or authenticated without the role — is turned away
// through the same 401 entry point.
func RequireRole(role model.RoleName, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authentication := FromRequest(r)
		if authentication == nil || !model.HasRole(authentication.Principal, role) {
			Commence(w)
			return
		}
		next(w, r)
	}
}
