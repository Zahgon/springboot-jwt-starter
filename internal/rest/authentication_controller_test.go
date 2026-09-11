package rest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
	"github.com/bfwg/springboot-jwt-starter/internal/apptest"
	"github.com/bfwg/springboot-jwt-starter/internal/clock"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

const authTestUsername = "testUser"

// stubUserDetailsService stands in for the mocked CustomUserDetailsService.
// It answers every lookup with a fixed user — deliberately named something
// other than the token subject, so token validation fails and no principal is
// established unless a test injects one.
type stubUserDetailsService struct{ user model.UserDetails }

func (s stubUserDetailsService) LoadUserByUsername(string) (model.UserDetails, error) {
	return s.user, nil
}

// newAuthFixture reproduces the original's @Before: a 100 second lifetime, a
// pinned secret, a scriptable clock, and a stubbed user-details lookup that
// returns a ROLE_USER account named "username" whose password was last reset
// yesterday.
func newAuthFixture(t *testing.T) (*app.App, *clock.Mock) {
	t.Helper()

	timeProviderMock := clock.NewMock(time.Now())
	user := &model.User{
		Username:              "username",
		LastPasswordResetDate: model.Timestamp(time.Now().AddDate(0, 0, -1)),
		Authorities:           []model.Authority{{ID: 0, Name: model.RoleUser}},
	}

	application := apptest.New(t, app.Options{
		Clock:       timeProviderMock,
		UserDetails: stubUserDetailsService{user: user},
	})

	application.TokenHelper.ExpiresIn = 100
	application.TokenHelper.Secret = "wb!.V.G]e&;4Q78&b,n[6F]PQg613!)-Gpy*CN{/[@LN.Z1Mn*aKH!*zg/pXLp/4"

	return application, timeProviderMock
}

// performWithHeader runs a request carrying an Authorization header.
func performWithHeader(t *testing.T, handler http.Handler, method, target, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

// assertJSONContent compares a response body against expected JSON, ignoring
// key order and formatting, which is what MockMvc's content().json() did.
func assertJSONContent(t *testing.T, body, expected string) {
	t.Helper()
	var got, want any
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("response body is not JSON: %v (body was %q)", err, body)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("expected value is not JSON: %v (was %q)", err, expected)
	}
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("content = %s, want %s", gotJSON, wantJSON)
	}
}

// jsonToken renders a token the way the original's expectation string did: a
// bare null when there is no token, and a quoted string otherwise.
func jsonToken(token string) string {
	if token == "" {
		return "null"
	}
	encoded, _ := json.Marshal(token)
	return string(encoded)
}

func TestShouldGetEmptyTokenStateWhenGivenValidOldToken(t *testing.T) {
	application, timeProviderMock := newAuthFixture(t)
	timeProviderMock.Returns(time.Now().AddDate(0, 0, -1))

	mvc := apptest.WithAnonymousUser(application.Handler)
	response := performWithHeader(t, mvc, http.MethodPost, "/auth/refresh", "Bearer 123")

	assertJSONContent(t, response.Body.String(), `{"access_token":null,"expires_in":null}`)
}

func TestShouldRefreshNotExpiredWebToken(t *testing.T) {
	application, timeProviderMock := newAuthFixture(t)
	timeProviderMock.Returns(time.UnixMilli(30))

	token := application.TokenHelper.GenerateToken(authTestUsername)
	refreshedToken := application.TokenHelper.RefreshToken(token)

	mvc := apptest.WithMockUser(application.Handler, model.RoleUser)
	response := performWithHeader(t, mvc, http.MethodPost, "/auth/refresh", "Bearer "+token)

	assertJSONContent(t, response.Body.String(),
		fmt.Sprintf(`{"access_token":%s,"expires_in":100}`, jsonToken(refreshedToken)))
}

func TestShouldNotRefreshExpiredWebToken(t *testing.T) {
	application, timeProviderMock := newAuthFixture(t)
	beforeSomeTime := time.Now().Add(-15 * time.Second)
	timeProviderMock.Returns(beforeSomeTime)

	token := application.TokenHelper.GenerateToken(authTestUsername)

	mvc := apptest.WithAnonymousUser(application.Handler)
	response := performWithHeader(t, mvc, http.MethodPost, "/auth/refresh", "Bearer "+token)

	assertJSONContent(t, response.Body.String(), `{"access_token":null,"expires_in":null}`)
}
