// Package apptest provides the wiring test packages need to exercise the
// application over HTTP: a wired app, and a way to run a request as an
// anonymous user or as a user holding a given role without logging in.
package apptest

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/security/auth"
)

// Assets returns the repository's web assets from disk, which keeps test
// wiring independent of the embedding in main.
func Assets(t *testing.T) fs.FS {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test source file")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	web := filepath.Join(root, "web")
	if _, err := os.Stat(web); err != nil {
		t.Fatalf("locating web assets: %v", err)
	}
	return os.DirFS(web)
}

// New wires an application for a test and closes it when the test ends. Each
// test gets its own in-memory database, named after the test.
func New(t *testing.T, opts app.Options) *app.App {
	t.Helper()
	if opts.DBName == "" {
		opts.DBName = t.Name()
	}
	application, err := app.New(Assets(t), opts)
	if err != nil {
		t.Fatalf("wiring the application: %v", err)
	}
	t.Cleanup(func() { _ = application.Close() })
	return application
}

// mockPrincipal builds a principal named "user" holding the given roles. It is
// a real model.User, so the role guard sees exactly what it sees in production.
func mockPrincipal(roles ...model.RoleName) *model.User {
	authorities := make([]model.Authority, 0, len(roles))
	for i, role := range roles {
		authorities = append(authorities, model.Authority{ID: int64(i), Name: role})
	}
	return &model.User{Username: "user", Enabled: true, Authorities: authorities}
}

// WithMockUser wraps a handler so every request through it carries a principal
// named "user" holding the given roles. It wraps the outside of the chain, so
// the token filter leaves the injected principal alone.
func WithMockUser(next http.Handler, roles ...model.RoleName) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, auth.WithAuthentication(r, &auth.TokenBasedAuthentication{
			Principal: mockPrincipal(roles...),
		}))
	})
}

// WithAnonymousUser wraps a handler so requests through it carry no principal.
func WithAnonymousUser(next http.Handler) http.Handler { return next }
