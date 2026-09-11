package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
	"github.com/bfwg/springboot-jwt-starter/internal/apptest"
)

// End-to-end coverage of the login/authorize flow. The original relied on
// Spring Security for all of this; the port implements it and so tests it.

func newLiveApp(t *testing.T) *app.App {
	t.Helper()
	return apptest.New(t, app.Options{})
}

// request runs a request with an optional JSON body and bearer token.
func request(t *testing.T, handler http.Handler, method, target, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	r := httptest.NewRequest(method, target, reader)
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

// login performs a real login and returns the issued token.
func login(t *testing.T, handler http.Handler, username, password string) string {
	t.Helper()
	response := request(t, handler, http.MethodPost, "/auth/login",
		`{"username":"`+username+`","password":"`+password+`"}`, "")
	if response.Code != http.StatusOK {
		t.Fatalf("login as %q: status = %d, want 200", username, response.Code)
	}
	var state struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &state); err != nil {
		t.Fatalf("decoding the login response: %v", err)
	}
	if state.AccessToken == "" {
		t.Fatal("login returned no access token")
	}
	return state.AccessToken
}

func TestLoginIssuesATokenForBothSeededAccounts(t *testing.T) {
	handler := newLiveApp(t).Handler

	for _, username := range []string{"user", "admin"} {
		t.Run(username, func(t *testing.T) {
			response := request(t, handler, http.MethodPost, "/auth/login",
				`{"username":"`+username+`","password":"123"}`, "")
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			var state struct {
				AccessToken string `json:"access_token"`
				ExpiresIn   int64  `json:"expires_in"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &state); err != nil {
				t.Fatalf("decoding: %v", err)
			}
			if strings.Count(state.AccessToken, ".") != 2 {
				t.Errorf("access_token %q is not a compact JWS", state.AccessToken)
			}
			if state.ExpiresIn != 300 {
				t.Errorf("expires_in = %d, want the configured 300", state.ExpiresIn)
			}
		})
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	handler := newLiveApp(t).Handler

	for name, body := range map[string]string{
		"wrong password": `{"username":"user","password":"wrong"}`,
		"unknown user":   `{"username":"nobody","password":"123"}`,
		"empty body":     `{}`,
		"malformed json": `{"username":`,
		"empty password": `{"username":"user","password":""}`,
	} {
		t.Run(name, func(t *testing.T) {
			response := request(t, handler, http.MethodPost, "/auth/login", body, "")
			if response.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", response.Code)
			}
			if response.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", response.Body.String())
			}
		})
	}
}

func TestWhoamiReturnsTheAuthenticatedAccount(t *testing.T) {
	handler := newLiveApp(t).Handler
	token := login(t, handler, "user", "123")

	response := request(t, handler, http.MethodGet, "/api/whoami", "", token)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	const want = `{"id":1,"username":"user","firstName":"Fan","lastName":"Jin",` +
		`"email":"user@example.com","phoneNumber":"+1234567890","enabled":true,` +
		`"lastPasswordResetDate":"2017-10-02T04:58:58.508+00:00",` +
		`"authorities":[{"authority":"ROLE_USER"}]}`
	if got := response.Body.String(); got != want {
		t.Errorf("whoami body\n got: %s\nwant: %s", got, want)
	}
}

func TestRoleGuardsSeparateUserFromAdmin(t *testing.T) {
	handler := newLiveApp(t).Handler
	userToken := login(t, handler, "user", "123")
	adminToken := login(t, handler, "admin", "123")

	for _, tc := range []struct {
		name   string
		target string
		token  string
		want   int
	}{
		{"user denied the roster", "/api/user/all", userToken, http.StatusUnauthorized},
		{"admin allowed the roster", "/api/user/all", adminToken, http.StatusOK},
		{"user denied a lookup", "/api/user/1", userToken, http.StatusUnauthorized},
		{"admin allowed a lookup", "/api/user/1", adminToken, http.StatusOK},
		{"user allowed whoami", "/api/whoami", userToken, http.StatusOK},
		{"admin allowed whoami", "/api/whoami", adminToken, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := request(t, handler, http.MethodGet, tc.target, "", tc.token).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestLoadAllReturnsBothAccountsOrderedByID(t *testing.T) {
	handler := newLiveApp(t).Handler
	adminToken := login(t, handler, "admin", "123")

	response := request(t, handler, http.MethodGet, "/api/user/all", "", adminToken)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	var users []struct {
		ID          int64 `json:"id"`
		Authorities []struct {
			Authority string `json:"authority"`
		} `json:"authorities"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &users); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d accounts, want 2", len(users))
	}
	if users[0].ID != 1 || users[1].ID != 2 {
		t.Errorf("ids = %d, %d; want 1, 2 in ascending order", users[0].ID, users[1].ID)
	}
	if len(users[1].Authorities) != 2 {
		t.Errorf("admin holds %d authorities, want 2", len(users[1].Authorities))
	}
}

func TestLookupOfAMissingAccountIs200WithAnEmptyBody(t *testing.T) {
	handler := newLiveApp(t).Handler
	adminToken := login(t, handler, "admin", "123")

	response := request(t, handler, http.MethodGet, "/api/user/999", "", adminToken)
	if response.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", response.Code)
	}
	if response.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", response.Body.String())
	}
}

func TestNonNumericUserIDIs401(t *testing.T) {
	handler := newLiveApp(t).Handler
	adminToken := login(t, handler, "admin", "123")

	response := request(t, handler, http.MethodGet, "/api/user/abc", "", adminToken)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", response.Code)
	}
}

func TestUnusableBearerValuesLeaveTheRequestAnonymous(t *testing.T) {
	handler := newLiveApp(t).Handler

	for name, header := range map[string]string{
		"garbage token":  "Bearer garbage",
		"empty token":    "Bearer ",
		"missing prefix": "some-token-without-a-prefix",
		"wrong scheme":   "Basic dXNlcjoxMjM=",
	} {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
			r.Header.Set("Authorization", header)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, r)
			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", recorder.Code)
			}
		})
	}
}

func TestRefreshWithoutAPrincipalIs202AndAllNull(t *testing.T) {
	handler := newLiveApp(t).Handler

	for name, token := range map[string]string{
		"no token":     "",
		"bogus bearer": "123",
	} {
		t.Run(name, func(t *testing.T) {
			response := request(t, handler, http.MethodPost, "/auth/refresh", "", token)
			if response.Code != http.StatusAccepted {
				t.Errorf("status = %d, want 202", response.Code)
			}
			if got, want := response.Body.String(), `{"access_token":null,"expires_in":null}`; got != want {
				t.Errorf("body = %s, want %s", got, want)
			}
		})
	}
}

func TestRefreshWithAPrincipalIssuesAUsableToken(t *testing.T) {
	handler := newLiveApp(t).Handler
	token := login(t, handler, "user", "123")

	response := request(t, handler, http.MethodPost, "/auth/refresh", "", token)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	var state struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &state); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if state.ExpiresIn != 300 {
		t.Errorf("expires_in = %d, want 300", state.ExpiresIn)
	}
	if request(t, handler, http.MethodGet, "/api/whoami", "", state.AccessToken).Code != http.StatusOK {
		t.Error("the refreshed token does not authenticate")
	}
}

// Changing a password stamps the reset instant, which by contract invalidates
// every token issued before it.
func TestChangePasswordRotatesCredentialsAndInvalidatesOldTokens(t *testing.T) {
	handler := newLiveApp(t).Handler
	token := login(t, handler, "user", "123")

	wrongOld := request(t, handler, http.MethodPost, "/auth/change-password",
		`{"oldPassword":"nope","newPassword":"x"}`, token)
	if wrongOld.Code != http.StatusUnauthorized {
		t.Errorf("wrong old password: status = %d, want 401", wrongOld.Code)
	}

	changed := request(t, handler, http.MethodPost, "/auth/change-password",
		`{"oldPassword":"123","newPassword":"456"}`, token)
	if changed.Code != http.StatusAccepted {
		t.Fatalf("change-password: status = %d, want 202", changed.Code)
	}
	if got, want := changed.Body.String(), `{"result":"success"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}

	if request(t, handler, http.MethodPost, "/auth/login",
		`{"username":"user","password":"456"}`, "").Code != http.StatusOK {
		t.Error("the new password does not authenticate")
	}
	if request(t, handler, http.MethodPost, "/auth/login",
		`{"username":"user","password":"123"}`, "").Code != http.StatusUnauthorized {
		t.Error("the old password still authenticates")
	}
	if got := request(t, handler, http.MethodGet, "/api/whoami", "", token).Code; got != http.StatusUnauthorized {
		t.Errorf("a token issued before the password change still works: status = %d", got)
	}
}

func TestChangePasswordRequiresAPrincipal(t *testing.T) {
	handler := newLiveApp(t).Handler

	response := request(t, handler, http.MethodPost, "/auth/change-password",
		`{"oldPassword":"123","newPassword":"456"}`, "")
	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", response.Code)
	}
}
