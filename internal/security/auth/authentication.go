// Package auth carries the authenticated principal through a request and
// decides how an unauthenticated one is turned away.
package auth

import (
	"context"
	"net/http"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

// contextKey is unexported so no other package can collide with it.
type contextKey struct{}

var principalKey contextKey

// TokenBasedAuthentication is the authenticated principal established from a
// bearer token, together with the token it came from.
type TokenBasedAuthentication struct {
	Principal model.UserDetails
	Token     string
}

// Name returns the authenticated username.
func (a *TokenBasedAuthentication) Name() string { return a.Principal.GetUsername() }

// WithAuthentication returns a request carrying the given principal.
func WithAuthentication(r *http.Request, a *TokenBasedAuthentication) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), principalKey, a))
}

// FromRequest returns the principal established for this request, or nil when
// the request is anonymous.
func FromRequest(r *http.Request) *TokenBasedAuthentication {
	a, _ := r.Context().Value(principalKey).(*TokenBasedAuthentication)
	return a
}
