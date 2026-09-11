// Package service holds the application services that sit between the HTTP
// controllers and the repository.
package service

import (
	"fmt"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/repository"
)

// ErrUsernameNotFound reports a username with no matching account.
type ErrUsernameNotFound struct{ Username string }

func (e *ErrUsernameNotFound) Error() string {
	return fmt.Sprintf("No user found with username '%s'.", e.Username)
}

// UserDetailsService resolves a username to a principal.
type UserDetailsService interface {
	LoadUserByUsername(username string) (model.UserDetails, error)
}

// CustomUserDetailsService loads principals from the user repository.
type CustomUserDetailsService struct {
	users *repository.UserRepository
}

// NewCustomUserDetailsService builds the service over a repository.
func NewCustomUserDetailsService(users *repository.UserRepository) *CustomUserDetailsService {
	return &CustomUserDetailsService{users: users}
}

// LoadUserByUsername returns the principal for a username, or
// ErrUsernameNotFound when there is no such account.
func (s *CustomUserDetailsService) LoadUserByUsername(username string) (model.UserDetails, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &ErrUsernameNotFound{Username: username}
	}
	return user, nil
}
