// Package crypto supplies the password hashing the application relied on
// Spring Security's BCryptPasswordEncoder for.
package crypto

import "golang.org/x/crypto/bcrypt"

// BCryptCost is the work factor BCryptPasswordEncoder uses when no strength is
// given. Hashes already in the database may carry a different cost and are
// still verified correctly, because the cost is encoded in the hash itself.
const BCryptCost = 10

// PasswordEncoder hashes and verifies passwords.
type PasswordEncoder interface {
	Encode(rawPassword string) (string, error)
	Matches(rawPassword, encodedPassword string) bool
}

// BCryptPasswordEncoder is the BCrypt implementation.
type BCryptPasswordEncoder struct{}

// NewBCryptPasswordEncoder returns a BCrypt-backed encoder.
func NewBCryptPasswordEncoder() BCryptPasswordEncoder { return BCryptPasswordEncoder{} }

// Encode hashes a raw password at the default cost.
func (BCryptPasswordEncoder) Encode(rawPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), BCryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Matches reports whether the raw password produced the stored hash.
func (BCryptPasswordEncoder) Matches(rawPassword, encodedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(rawPassword)) == nil
}
