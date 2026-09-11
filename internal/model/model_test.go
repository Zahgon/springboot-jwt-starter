package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

// The JSON encodings below were previously produced by Jackson and are part of
// the wire contract, so they are asserted byte for byte.

func TestUserJSONIsByteExact(t *testing.T) {
	resetAt, err := model.ParseTimestamp("2017-10-01 21:58:58.508-07")
	if err != nil {
		t.Fatalf("parsing the seeded instant: %v", err)
	}

	user := &model.User{
		ID:                    1,
		Username:              "user",
		Password:              "$2a$04$Vbug2lwwJGrvUXTj6z7ff.97IzVBkrJ1XfApfGNl.Z695zqcnPYra",
		FirstName:             "Fan",
		LastName:              "Jin",
		Email:                 "user@example.com",
		PhoneNumber:           "+1234567890",
		Enabled:               true,
		LastPasswordResetDate: resetAt,
		Authorities:           []model.Authority{{ID: 1, Name: model.RoleUser}},
	}

	encoded, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshalling user: %v", err)
	}

	const want = `{"id":1,"username":"user","firstName":"Fan","lastName":"Jin",` +
		`"email":"user@example.com","phoneNumber":"+1234567890","enabled":true,` +
		`"lastPasswordResetDate":"2017-10-02T04:58:58.508+00:00",` +
		`"authorities":[{"authority":"ROLE_USER"}]}`

	if string(encoded) != want {
		t.Errorf("user JSON\n got: %s\nwant: %s", encoded, want)
	}
}

func TestAuthorityJSONExposesOnlyTheRoleName(t *testing.T) {
	encoded, err := json.Marshal(model.Authority{ID: 7, Name: model.RoleAdmin})
	if err != nil {
		t.Fatalf("marshalling authority: %v", err)
	}
	if got, want := string(encoded), `{"authority":"ROLE_ADMIN"}`; got != want {
		t.Errorf("authority JSON = %s, want %s", got, want)
	}
}

func TestTimestampRendersUTCWithMillisecondsAndOffset(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{"seeded user", "2017-10-01 21:58:58.508-07", `"2017-10-02T04:58:58.508+00:00"`},
		{"seeded admin", "2017-10-01 18:57:58.508-07", `"2017-10-02T01:57:58.508+00:00"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts, err := model.ParseTimestamp(tc.input)
			if err != nil {
				t.Fatalf("parsing %q: %v", tc.input, err)
			}
			encoded, err := json.Marshal(ts)
			if err != nil {
				t.Fatalf("marshalling: %v", err)
			}
			if string(encoded) != tc.want {
				t.Errorf("timestamp JSON = %s, want %s", encoded, tc.want)
			}
		})
	}
}

func TestTimestampRoundTripsThroughStorage(t *testing.T) {
	original := model.Timestamp(time.Date(2021, 3, 4, 5, 6, 7, 890_000_000, time.UTC))

	restored, err := model.ParseTimestamp(original.String())
	if err != nil {
		t.Fatalf("re-parsing %q: %v", original.String(), err)
	}
	if !restored.Time().Equal(original.Time()) {
		t.Errorf("round trip gave %v, want %v", restored.Time(), original.Time())
	}
}

func TestUserTokenStateJSON(t *testing.T) {
	populated, err := json.Marshal(model.NewUserTokenState("abc", 300))
	if err != nil {
		t.Fatalf("marshalling populated state: %v", err)
	}
	if got, want := string(populated), `{"access_token":"abc","expires_in":300}`; got != want {
		t.Errorf("populated token state = %s, want %s", got, want)
	}

	empty, err := json.Marshal(model.EmptyUserTokenState())
	if err != nil {
		t.Fatalf("marshalling empty state: %v", err)
	}
	if got, want := string(empty), `{"access_token":null,"expires_in":null}`; got != want {
		t.Errorf("empty token state = %s, want %s", got, want)
	}

	// Refreshing for a known principal reports the expiry even when no token
	// could be minted.
	partial, err := json.Marshal(model.NewUserTokenStateWithExpiry("", 100))
	if err != nil {
		t.Fatalf("marshalling partial state: %v", err)
	}
	if got, want := string(partial), `{"access_token":null,"expires_in":100}`; got != want {
		t.Errorf("partial token state = %s, want %s", got, want)
	}
}

func TestSetPasswordStampsTheResetInstant(t *testing.T) {
	user := &model.User{}
	instant := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	user.SetPassword("hash", instant)

	if user.Password != "hash" {
		t.Errorf("password = %q, want %q", user.Password, "hash")
	}
	if !user.GetLastPasswordResetDate().Time().Equal(instant) {
		t.Errorf("reset instant = %v, want %v", user.GetLastPasswordResetDate().Time(), instant)
	}
}

func TestHasRole(t *testing.T) {
	user := &model.User{Authorities: []model.Authority{
		{Name: model.RoleUser}, {Name: model.RoleAdmin},
	}}

	if !model.HasRole(user, model.RoleAdmin) {
		t.Error("expected the admin role to be held")
	}
	if model.HasRole(&model.User{}, model.RoleUser) {
		t.Error("a user with no authorities holds no role")
	}
	if model.HasRole(nil, model.RoleUser) {
		t.Error("a nil principal holds no role")
	}
}

// The four status flags are constant in this application, but they are part of
// the UserDetails contract the security layer depends on.
func TestUserAccountStatusFlags(t *testing.T) {
	user := &model.User{Enabled: true}

	if !user.IsEnabled() {
		t.Error("IsEnabled must report the Enabled field")
	}
	if !user.IsAccountNonExpired() {
		t.Error("IsAccountNonExpired must be true")
	}
	if !user.IsAccountNonLocked() {
		t.Error("IsAccountNonLocked must be true")
	}
	if !user.IsCredentialsNonExpired() {
		t.Error("IsCredentialsNonExpired must be true")
	}
	if (&model.User{Enabled: false}).IsEnabled() {
		t.Error("a disabled account must not report itself enabled")
	}
}

func TestUserSatisfiesUserDetails(t *testing.T) {
	var details model.UserDetails = &model.User{
		Username:    "u",
		Authorities: []model.Authority{{Name: model.RoleAdmin}},
	}

	if details.GetUsername() != "u" {
		t.Errorf("GetUsername = %q, want %q", details.GetUsername(), "u")
	}
	if got := details.GetAuthorities(); len(got) != 1 || got[0].Name != model.RoleAdmin {
		t.Errorf("GetAuthorities = %v, want [ROLE_ADMIN]", got)
	}
}
