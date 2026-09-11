// Package security mints and validates the JSON Web Tokens that authenticate
// every non-public request.
package security

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/bfwg/springboot-jwt-starter/internal/clock"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

// bearerPrefix is the exact prefix an Authorization header must carry.
const bearerPrefix = "Bearer "

// signingMethod is HMAC-SHA512, matching the original's Jwts.SIG.HS512.
var signingMethod = jwt.SigningMethodHS512

// Claims is the token payload. The field order below is the JSON key order in
// the encoded payload, and the set is closed: the original emits iss, sub, iat
// and exp, and nothing else.
type Claims struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	IssuedAt int64  `json:"iat"`
	Expiry   int64  `json:"exp"`
}

// GetExpirationTime implements jwt.Claims, which is what drives expiry checking.
func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.Expiry, 0)), nil
}

// GetIssuedAt implements jwt.Claims.
func (c Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.IssuedAt, 0)), nil
}

// GetNotBefore implements jwt.Claims. The original never sets nbf.
func (c Claims) GetNotBefore() (*jwt.NumericDate, error) { return nil, nil }

// GetIssuer implements jwt.Claims.
func (c Claims) GetIssuer() (string, error) { return c.Issuer, nil }

// GetSubject implements jwt.Claims.
func (c Claims) GetSubject() (string, error) { return c.Subject, nil }

// GetAudience implements jwt.Claims. The original never sets aud.
func (c Claims) GetAudience() (jwt.ClaimStrings, error) { return nil, nil }

// TokenHelper issues and inspects tokens. AppName, Secret, ExpiresIn and
// AuthHeader come from configuration and are settable directly so tests can
// pin them, as the original's tests did through ReflectionTestUtils.
type TokenHelper struct {
	AppName    string
	Secret     string
	ExpiresIn  int
	AuthHeader string
	Time       clock.Provider
}

// NewTokenHelper builds a helper from configuration values.
func NewTokenHelper(appName, secret string, expiresIn int, authHeader string, timeProvider clock.Provider) *TokenHelper {
	return &TokenHelper{
		AppName:    appName,
		Secret:     secret,
		ExpiresIn:  expiresIn,
		AuthHeader: authHeader,
		Time:       timeProvider,
	}
}

// key is the raw UTF-8 bytes of the configured secret, exactly as
// Keys.hmacShaKeyFor(SECRET.getBytes()) produced.
func (h *TokenHelper) key() []byte { return []byte(h.Secret) }

// sign encodes and signs a payload, emitting a bare {"alg":"HS512"} header with
// no "typ" member, which is what the original produces.
func (h *TokenHelper) sign(claims Claims) string {
	token := jwt.NewWithClaims(signingMethod, claims)
	delete(token.Header, "typ")
	signed, err := token.SignedString(h.key())
	if err != nil {
		return ""
	}
	return signed
}

// GenerateToken mints a token for the given username.
//
// The clock is read twice — once for iat and once as the base for exp — which
// mirrors the original and is observable through an injected time source.
func (h *TokenHelper) GenerateToken(username string) string {
	issuedAt := h.Time.Now()
	return h.sign(Claims{
		Issuer:   h.AppName,
		Subject:  username,
		IssuedAt: issuedAt.Unix(),
		Expiry:   h.generateExpirationDate().Unix(),
	})
}

// RefreshToken re-issues a token from its existing claims, so the issuer and
// subject carry over while iat and exp are renewed. A token that cannot be
// parsed — including one that has expired — yields the empty string rather than
// an error.
func (h *TokenHelper) RefreshToken(token string) string {
	now := h.Time.Now()
	claims := h.getAllClaimsFromToken(token)
	if claims == nil {
		return ""
	}
	claims.IssuedAt = now.Unix()
	claims.Expiry = h.generateExpirationDate().Unix()
	return h.sign(*claims)
}

// GetUsernameFromToken returns the token's subject, or "" if it cannot be read.
func (h *TokenHelper) GetUsernameFromToken(token string) string {
	claims := h.getAllClaimsFromToken(token)
	if claims == nil {
		return ""
	}
	return claims.Subject
}

// GetIssuedAtDateFromToken returns the token's issue instant, or the zero time
// if it cannot be read.
func (h *TokenHelper) GetIssuedAtDateFromToken(token string) time.Time {
	claims := h.getAllClaimsFromToken(token)
	if claims == nil {
		return time.Time{}
	}
	return time.Unix(claims.IssuedAt, 0)
}

// getAllClaimsFromToken verifies the signature and the expiry, returning nil on
// any failure.
func (h *TokenHelper) getAllClaimsFromToken(token string) *Claims {
	parsed, err := jwt.ParseWithClaims(token, &Claims{},
		func(*jwt.Token) (any, error) { return h.key(), nil },
		jwt.WithValidMethods([]string{signingMethod.Alg()}),
	)
	if err != nil || !parsed.Valid {
		return nil
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// generateExpirationDate reads the clock again and adds the configured lifetime.
func (h *TokenHelper) generateExpirationDate() time.Time {
	return h.Time.Now().Add(time.Duration(h.ExpiresIn) * time.Second)
}

// GetExpiredIn returns the configured token lifetime in seconds.
func (h *TokenHelper) GetExpiredIn() int { return h.ExpiresIn }

// ValidateToken reports whether the token authenticates the given principal.
// Both conditions must hold: the subject names that principal, and the token
// was issued no earlier than the principal's last password reset.
func (h *TokenHelper) ValidateToken(token string, userDetails model.UserDetails) bool {
	username := h.GetUsernameFromToken(token)
	created := h.GetIssuedAtDateFromToken(token)
	return username != "" && username == userDetails.GetUsername() &&
		!isCreatedBeforeLastPasswordReset(created, userDetails.GetLastPasswordResetDate())
}

func isCreatedBeforeLastPasswordReset(created time.Time, lastPasswordReset model.Timestamp) bool {
	return !lastPasswordReset.IsZero() && created.Before(lastPasswordReset.Time())
}

// GetToken extracts the bearer token from the configured header, or "" when the
// header is absent or does not carry the exact "Bearer " prefix.
func (h *TokenHelper) GetToken(r *http.Request) string {
	authHeader := h.GetAuthHeaderFromHeader(r)
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return authHeader[len(bearerPrefix):]
	}
	return ""
}

// GetAuthHeaderFromHeader returns the raw value of the configured auth header.
func (h *TokenHelper) GetAuthHeaderFromHeader(r *http.Request) string {
	return r.Header.Get(h.AuthHeader)
}
