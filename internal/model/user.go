package model

import "time"

// UserDetails is the contract the security layer needs from a principal. It is
// satisfied by User in production and by lightweight doubles in tests.
type UserDetails interface {
	GetUsername() string
	GetAuthorities() []Authority
	GetLastPasswordResetDate() Timestamp
	IsEnabled() bool
	IsAccountNonExpired() bool
	IsAccountNonLocked() bool
	IsCredentialsNonExpired() bool
}

// User is an account. Field order below is the JSON key order on the wire; the
// password hash and the three account-status flags are never serialized.
type User struct {
	ID                    int64       `json:"id"`
	Username              string      `json:"username"`
	Password              string      `json:"-"`
	FirstName             string      `json:"firstName"`
	LastName              string      `json:"lastName"`
	Email                 string      `json:"email"`
	PhoneNumber           string      `json:"phoneNumber"`
	Enabled               bool        `json:"enabled"`
	LastPasswordResetDate Timestamp   `json:"lastPasswordResetDate"`
	Authorities           []Authority `json:"authorities"`
}

// GetUsername returns the account's username.
func (u *User) GetUsername() string { return u.Username }

// GetPassword returns the stored password hash.
func (u *User) GetPassword() string { return u.Password }

// GetAuthorities returns the roles granted to the account.
func (u *User) GetAuthorities() []Authority { return u.Authorities }

// GetLastPasswordResetDate returns the instant the password last changed.
func (u *User) GetLastPasswordResetDate() Timestamp { return u.LastPasswordResetDate }

// SetPassword stores a new password hash and, as a side effect, stamps the
// password-reset instant to now — which invalidates every token issued earlier.
func (u *User) SetPassword(hash string, now time.Time) {
	u.LastPasswordResetDate = Timestamp(now)
	u.Password = hash
}

// IsEnabled reports whether the account is active.
func (u *User) IsEnabled() bool { return u.Enabled }

// IsAccountNonExpired is always true for this application.
func (u *User) IsAccountNonExpired() bool { return true }

// IsAccountNonLocked is always true for this application.
func (u *User) IsAccountNonLocked() bool { return true }

// IsCredentialsNonExpired is always true for this application.
func (u *User) IsCredentialsNonExpired() bool { return true }

// HasRole reports whether the principal holds the named authority.
func HasRole(d UserDetails, role RoleName) bool {
	if d == nil {
		return false
	}
	for _, a := range d.GetAuthorities() {
		if a.Name == role {
			return true
		}
	}
	return false
}
