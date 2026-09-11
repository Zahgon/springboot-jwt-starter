package security_test

import "github.com/bfwg/springboot-jwt-starter/internal/model"

// userDetailsDummy is a principal carrying only a username, with every status
// flag false and no authorities.
type userDetailsDummy struct {
	username string
}

func newUserDetailsDummy(username string) *userDetailsDummy {
	return &userDetailsDummy{username: username}
}

func (d *userDetailsDummy) GetUsername() string                       { return d.username }
func (d *userDetailsDummy) GetAuthorities() []model.Authority         { return nil }
func (d *userDetailsDummy) GetLastPasswordResetDate() model.Timestamp { return model.Timestamp{} }
func (d *userDetailsDummy) IsEnabled() bool                           { return false }
func (d *userDetailsDummy) IsAccountNonExpired() bool                 { return false }
func (d *userDetailsDummy) IsAccountNonLocked() bool                  { return false }
func (d *userDetailsDummy) IsCredentialsNonExpired() bool             { return false }

// mockUser is a principal whose username and last-password-reset instant are
// set per test, standing in for the mocked User the original used.
type mockUser struct {
	username              string
	lastPasswordResetDate model.Timestamp
}

func (m *mockUser) GetUsername() string                       { return m.username }
func (m *mockUser) GetAuthorities() []model.Authority         { return nil }
func (m *mockUser) GetLastPasswordResetDate() model.Timestamp { return m.lastPasswordResetDate }
func (m *mockUser) IsEnabled() bool                           { return true }
func (m *mockUser) IsAccountNonExpired() bool                 { return true }
func (m *mockUser) IsAccountNonLocked() bool                  { return true }
func (m *mockUser) IsCredentialsNonExpired() bool             { return true }
