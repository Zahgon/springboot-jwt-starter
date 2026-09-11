package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
	"github.com/bfwg/springboot-jwt-starter/internal/apptest"
)

// The asset serving and the uniform failure surface below were previously
// supplied by Spring Security and by Maven's webjar packaging, so the port owns
// them now and they need tests of their own.

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	return apptest.New(t, app.Options{}).Handler
}

func get(t *testing.T, handler http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestServesTheApplicationsOwnStaticAssets(t *testing.T) {
	handler := newHandler(t)

	for _, target := range []string{
		"/", "/index.html", "/app.js",
		"/login/login.html", "/login/login.js",
		"/dashboard/dashboard.html", "/dashboard/dashboard.js",
		"/services/auth.js",
	} {
		t.Run(target, func(t *testing.T) {
			response := get(t, handler, target)
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			if response.Body.Len() == 0 {
				t.Error("asset was served empty")
			}
		})
	}
}

func TestRootServesIndex(t *testing.T) {
	handler := newHandler(t)

	if root, index := get(t, handler, "/").Body.String(), get(t, handler, "/index.html").Body.String(); root != index {
		t.Error("/ and /index.html must serve the same document")
	}
}

// DELIBERATE DEVIATION from the original, recorded in truth.md.
//
// The original serves four browser libraries under /webjars/, supplied by Maven
// as packaged jars. They are not vendored into this repository, so the whole
// /webjars space is absent and answers 401 like any other unknown path. The
// bundled UI cannot load its libraries as a result.
func TestWebjarsAreNotServed(t *testing.T) {
	handler := newHandler(t)

	for _, target := range []string{
		"/webjars/jquery/2.1.1/jquery.js",
		"/webjars/angularjs/1.5.8/angular.js",
		"/webjars/angular-route/1.5.9/angular-route.js",
		"/webjars/bootstrap/3.4.1/css/bootstrap.css",
		"/webjars/",
	} {
		t.Run(target, func(t *testing.T) {
			response := get(t, handler, target)
			if response.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", response.Code)
			}
			if response.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", response.Body.String())
			}
		})
	}
}

// index.html asks for bootstrap 3.3.7 while the original ships 3.4.1, so the
// original answers 401 for these two URLs. This port answers 401 as well —
// upstream because the version is wrong, here because no webjar is served at
// all. index.html is carried over unchanged either way: repointing it would be
// a bug fix, not a migration.
func TestBootstrap337IsAbsentJustAsUpstream(t *testing.T) {
	handler := newHandler(t)

	for _, target := range []string{
		"/webjars/bootstrap/3.3.7/css/bootstrap.css",
		"/webjars/bootstrap/3.3.7/js/bootstrap.js",
	} {
		t.Run(target, func(t *testing.T) {
			if got := get(t, handler, target).Code; got != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", got)
			}
		})
	}
}

// Every rejection leaves through the same door: 401 and an empty body. There is
// no 403 and no 404 anywhere in this application.
func TestFailureSurfaceIsAlways401WithAnEmptyBody(t *testing.T) {
	handler := newHandler(t)

	for _, target := range []string{
		"/api/whoami",   // guarded, no principal
		"/api/user/all", // guarded, no principal
		"/api/user/1",   // guarded, no principal
		"/nope",         // unknown path
		"/favicon.ico",  // permitted pattern, missing file
		"/webjars/",     // a directory, not a file
		"/user",         // unknown path
		"/login",        // a directory, not a file
	} {
		t.Run(target, func(t *testing.T) {
			response := get(t, handler, target)
			if response.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", response.Code)
			}
			if response.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", response.Body.String())
			}
		})
	}
}

func TestWrongMethodOnAKnownPathIsAlso401(t *testing.T) {
	handler := newHandler(t)

	for _, tc := range []struct{ method, target string }{
		{http.MethodGet, "/auth/login"},
		{http.MethodGet, "/auth/refresh"},
		{http.MethodPost, "/api/whoami"},
		{http.MethodPost, "/index.html"},
		{http.MethodDelete, "/api/user/all"},
	} {
		t.Run(tc.method+" "+tc.target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.target, nil))
			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", recorder.Code)
			}
			if recorder.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", recorder.Body.String())
			}
		})
	}
}

// The original serves static resources without a charset parameter; Go's own
// extension lookup would append one, so the media type is pinned here.
func TestStaticContentTypesCarryNoCharset(t *testing.T) {
	handler := newHandler(t)

	for target, want := range map[string]string{
		"/index.html":             "text/html",
		"/login/login.html":       "text/html",
		"/app.js":                 "text/javascript",
		"/services/auth.js":       "text/javascript",
		"/dashboard/dashboard.js": "text/javascript",
	} {
		t.Run(target, func(t *testing.T) {
			got := get(t, handler, target).Header().Get("Content-Type")
			if got != want {
				t.Errorf("Content-Type = %q, want %q", got, want)
			}
		})
	}
}
