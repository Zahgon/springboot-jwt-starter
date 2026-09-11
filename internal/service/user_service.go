package service

import (
	"errors"

	"github.com/bfwg/springboot-jwt-starter/internal/clock"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/repository"
	"github.com/bfwg/springboot-jwt-starter/internal/security/crypto"
)

// ErrBadCredentials reports a username/password pair that does not authenticate.
var ErrBadCredentials = errors.New("Bad credentials")

// UserService is the account-facing application service.
type UserService interface {
	FindByID(id int64) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindAll() ([]*model.User, error)
	Authenticate(username, password string) (*model.User, error)
	ChangePassword(username, oldPassword, newPassword string) error
}

// UserServiceImpl is the repository-backed implementation.
type UserServiceImpl struct {
	users    *repository.UserRepository
	encoder  crypto.PasswordEncoder
	details  UserDetailsService
	timeFunc clock.Provider
}

// NewUserService builds the account service.
func NewUserService(
	users *repository.UserRepository,
	encoder crypto.PasswordEncoder,
	details UserDetailsService,
	timeProvider clock.Provider,
) *UserServiceImpl {
	return &UserServiceImpl{users: users, encoder: encoder, details: details, timeFunc: timeProvider}
}

// FindByUsername returns the account with the given username, or nil.
func (s *UserServiceImpl) FindByUsername(username string) (*model.User, error) {
	return s.users.FindByUsername(username)
}

// FindByID returns the account with the given id, or nil.
func (s *UserServiceImpl) FindByID(id int64) (*model.User, error) {
	return s.users.FindByID(id)
}

// FindAll returns every account, ordered by id.
func (s *UserServiceImpl) FindAll() ([]*model.User, error) { return s.users.FindAll() }

// Authenticate verifies a username and password, returning the account on
// success and ErrBadCredentials otherwise. An unknown username and a wrong
// password are deliberately indistinguishable to the caller.
func (s *UserServiceImpl) Authenticate(username, password string) (*model.User, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil || !s.encoder.Matches(password, user.GetPassword()) {
		return nil, ErrBadCredentials
	}
	return user, nil
}

// ChangePassword re-authenticates the caller with the old password before
// storing the new one. Storing a password also stamps the password-reset
// instant, which invalidates every token issued earlier.
func (s *UserServiceImpl) ChangePassword(username, oldPassword, newPassword string) error {
	if _, err := s.Authenticate(username, oldPassword); err != nil {
		return err
	}

	details, err := s.details.LoadUserByUsername(username)
	if err != nil {
		return err
	}
	user, ok := details.(*model.User)
	if !ok {
		return ErrBadCredentials
	}

	hash, err := s.encoder.Encode(newPassword)
	if err != nil {
		return err
	}
	user.SetPassword(hash, s.timeFunc.Now())
	return s.users.Save(user)
}
