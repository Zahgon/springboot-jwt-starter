package service_test

import (
	"errors"
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/db"
	"github.com/bfwg/springboot-jwt-starter/internal/repository"
	"github.com/bfwg/springboot-jwt-starter/internal/service"
)

func newUserDetailsService(t *testing.T) *service.CustomUserDetailsService {
	t.Helper()
	database, err := db.Open(t.Name())
	if err != nil {
		t.Fatalf("opening the database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return service.NewCustomUserDetailsService(repository.New(database))
}

func TestLoadUserByUsername(t *testing.T) {
	details, err := newUserDetailsService(t).LoadUserByUsername("user")
	if err != nil {
		t.Fatalf("loading a known account: %v", err)
	}
	if details.GetUsername() != "user" {
		t.Errorf("username = %q, want %q", details.GetUsername(), "user")
	}
}

// The original raised UsernameNotFoundException with this exact wording.
func TestLoadUserByUsernameReportsAMissingAccount(t *testing.T) {
	_, err := newUserDetailsService(t).LoadUserByUsername("nobody")
	if err == nil {
		t.Fatal("expected an error for a missing account")
	}

	var notFound *service.ErrUsernameNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("error is %T, want *service.ErrUsernameNotFound", err)
	}
	if got, want := err.Error(), `No user found with username 'nobody'.`; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}
