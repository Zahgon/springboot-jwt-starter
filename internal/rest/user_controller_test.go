package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
	"github.com/bfwg/springboot-jwt-starter/internal/apptest"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

// perform runs a request through the handler and returns the recorded response.
func perform(t *testing.T, handler http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
	return recorder
}

func assertStatusIn2xx(t *testing.T, got int) {
	t.Helper()
	if got < 200 || got > 299 {
		t.Errorf("status = %d, want 2xx", got)
	}
}

func assertStatusIn4xx(t *testing.T, got int) {
	t.Helper()
	if got < 400 || got > 499 {
		t.Errorf("status = %d, want 4xx", got)
	}
}

func TestShouldGetUnauthorizedWithoutRole(t *testing.T) {
	application := apptest.New(t, app.Options{})
	mvc := apptest.WithAnonymousUser(application.Handler)

	if got := perform(t, mvc, http.MethodGet, "/user").Code; got != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestGetPersonsSuccessfullyWithUserRole(t *testing.T) {
	application := apptest.New(t, app.Options{})
	mvc := apptest.WithMockUser(application.Handler, model.RoleUser)

	assertStatusIn2xx(t, perform(t, mvc, http.MethodGet, "/api/whoami").Code)
}

func TestGetPersonsFailWithAnonymousUser(t *testing.T) {
	application := apptest.New(t, app.Options{})
	mvc := apptest.WithAnonymousUser(application.Handler)

	assertStatusIn4xx(t, perform(t, mvc, http.MethodGet, "/api/whoami").Code)
}

func TestGetAllUserSuccessWithAdminRole(t *testing.T) {
	application := apptest.New(t, app.Options{})
	mvc := apptest.WithMockUser(application.Handler, model.RoleAdmin)

	assertStatusIn2xx(t, perform(t, mvc, http.MethodGet, "/api/user/all").Code)
}

func TestGetAllUserFailWithUserRole(t *testing.T) {
	application := apptest.New(t, app.Options{})
	mvc := apptest.WithMockUser(application.Handler, model.RoleUser)

	assertStatusIn4xx(t, perform(t, mvc, http.MethodGet, "/api/user/all").Code)
}
