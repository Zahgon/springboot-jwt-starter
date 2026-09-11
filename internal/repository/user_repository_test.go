package repository_test

import (
	"testing"
	"time"

	"github.com/bfwg/springboot-jwt-starter/internal/db"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/repository"
)

// Persistence was Spring Data JPA's job in the original; the port writes the
// queries itself, so the seed loading and each query are tested here.

func newRepository(t *testing.T) *repository.UserRepository {
	t.Helper()
	database, err := db.Open(t.Name())
	if err != nil {
		t.Fatalf("opening the database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return repository.New(database)
}

func TestSeedDataIsLoaded(t *testing.T) {
	users, err := newRepository(t).FindAll()
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d seeded accounts, want 2", len(users))
	}

	first, second := users[0], users[1]
	if first.ID != 1 || second.ID != 2 {
		t.Errorf("ids = %d, %d; want them ascending from 1", first.ID, second.ID)
	}
	if first.Username != "user" || second.Username != "admin" {
		t.Errorf("usernames = %q, %q; want \"user\", \"admin\"", first.Username, second.Username)
	}
	if first.FirstName != "Fan" || first.LastName != "Jin" {
		t.Errorf("user name = %q %q, want Fan Jin", first.FirstName, first.LastName)
	}
	if first.Email != "user@example.com" || first.PhoneNumber != "+1234567890" {
		t.Errorf("user contact = %q / %q", first.Email, first.PhoneNumber)
	}
	if !first.Enabled || !second.Enabled {
		t.Error("both seeded accounts must be enabled")
	}
}

func TestSeededAuthorities(t *testing.T) {
	users, err := newRepository(t).FindAll()
	if err != nil {
		t.Fatalf("listing users: %v", err)
	}

	if got := users[0].GetAuthorities(); len(got) != 1 || got[0].Name != model.RoleUser {
		t.Errorf("user authorities = %v, want [ROLE_USER]", got)
	}
	got := users[1].GetAuthorities()
	if len(got) != 2 || got[0].Name != model.RoleUser || got[1].Name != model.RoleAdmin {
		t.Errorf("admin authorities = %v, want [ROLE_USER ROLE_ADMIN]", got)
	}
}

func TestSeededPasswordResetInstants(t *testing.T) {
	users := newRepository(t)

	for username, want := range map[string]string{
		"user":  "2017-10-02T04:58:58.508+00:00",
		"admin": "2017-10-02T01:57:58.508+00:00",
	} {
		t.Run(username, func(t *testing.T) {
			u, err := users.FindByUsername(username)
			if err != nil {
				t.Fatalf("lookup: %v", err)
			}
			encoded, err := u.GetLastPasswordResetDate().MarshalJSON()
			if err != nil {
				t.Fatalf("marshalling: %v", err)
			}
			if got := string(encoded); got != `"`+want+`"` {
				t.Errorf("reset instant = %s, want %q", got, want)
			}
		})
	}
}

func TestFindByUsernameAndID(t *testing.T) {
	users := newRepository(t)

	found, err := users.FindByUsername("admin")
	if err != nil {
		t.Fatalf("lookup by username: %v", err)
	}
	if found == nil || found.ID != 2 {
		t.Fatalf("FindByUsername(admin) = %v, want the account with id 2", found)
	}

	byID, err := users.FindByID(2)
	if err != nil {
		t.Fatalf("lookup by id: %v", err)
	}
	if byID == nil || byID.Username != "admin" {
		t.Fatalf("FindByID(2) = %v, want admin", byID)
	}
}

func TestMissingLookupsYieldNilWithoutError(t *testing.T) {
	users := newRepository(t)

	byName, err := users.FindByUsername("nobody")
	if err != nil {
		t.Errorf("FindByUsername on a missing account errored: %v", err)
	}
	if byName != nil {
		t.Errorf("FindByUsername(nobody) = %v, want nil", byName)
	}

	byID, err := users.FindByID(999)
	if err != nil {
		t.Errorf("FindByID on a missing account errored: %v", err)
	}
	if byID != nil {
		t.Errorf("FindByID(999) = %v, want nil", byID)
	}
}

func TestSavePersistsAPasswordChange(t *testing.T) {
	users := newRepository(t)

	user, err := users.FindByUsername("user")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	instant := time.Date(2024, 5, 6, 7, 8, 9, 250_000_000, time.UTC)
	user.SetPassword("$2a$10$newhashvalue", instant)

	if err := users.Save(user); err != nil {
		t.Fatalf("saving: %v", err)
	}

	reloaded, err := users.FindByUsername("user")
	if err != nil {
		t.Fatalf("re-reading: %v", err)
	}
	if reloaded.Password != "$2a$10$newhashvalue" {
		t.Errorf("password = %q, want the saved hash", reloaded.Password)
	}
	if !reloaded.GetLastPasswordResetDate().Time().Equal(instant) {
		t.Errorf("reset instant = %v, want %v", reloaded.GetLastPasswordResetDate().Time(), instant)
	}
}
