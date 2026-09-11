// Package server wires the routes, their guards and the static asset handler
// into a single http.Handler.
package server

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
	"github.com/bfwg/springboot-jwt-starter/internal/rest"
	"github.com/bfwg/springboot-jwt-starter/internal/security/auth"
)

// Handler is the application's HTTP surface.
type Handler struct {
	mux http.Handler
}

// ServeHTTP dispatches a request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

// New builds the routing table.
//
// Guarding is uniform: an endpoint either carries a role guard or is public,
// and every rejection — wrong role, no principal, wrong method, unknown path,
// missing asset — leaves through auth.Commence as 401 with an empty body.
func New(
	assets fs.FS,
	authentication *rest.AuthenticationController,
	users *rest.UserController,
	tokenFilter func(http.Handler) http.Handler,
) *Handler {
	mux := http.NewServeMux()

	// Public: credentials are the only thing that gets you in.
	mux.HandleFunc("/auth/login", methodGuard(http.MethodPost, authentication.CreateAuthenticationToken))
	// Public route, but its body depends on whether a principal was established.
	mux.HandleFunc("/auth/refresh", methodGuard(http.MethodPost, authentication.RefreshAuthenticationToken))
	mux.HandleFunc("/auth/change-password", methodGuard(http.MethodPost,
		auth.RequireRole(model.RoleUser, authentication.ChangePassword)))

	mux.HandleFunc("/api/whoami", methodGuard(http.MethodGet,
		auth.RequireRole(model.RoleUser, users.User)))
	// Registered before the parameterised pattern so "all" is not read as an id.
	mux.HandleFunc("/api/user/all", methodGuard(http.MethodGet,
		auth.RequireRole(model.RoleAdmin, users.LoadAll)))
	mux.HandleFunc("/api/user/{userId}", methodGuard(http.MethodGet,
		auth.RequireRole(model.RoleAdmin, users.LoadByID)))

	// Everything else is a static asset lookup, and a miss is a 401.
	mux.HandleFunc("/", staticHandler(assets))

	return &Handler{mux: tokenFilter(mux)}
}

// methodGuard rejects a request that reached a known path with the wrong method.
func methodGuard(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			auth.Commence(w)
			return
		}
		next(w, r)
	}
}

// assetPath maps a request path onto a location in the bundled asset tree.
//
// The original exposed a second resource root at "/webjars/...", backed by
// browser libraries that Maven supplied as packaged jars. Those are not
// vendored here (see truth.md), so there is one root: the application's own
// files, served from "/".
func assetPath(urlPath string) string {
	name := strings.TrimPrefix(path.Clean(urlPath), "/")
	if name == "" || name == "." {
		return "static/index.html"
	}
	return "static/" + name
}

// staticContentTypes are the media types the original serves for the asset
// extensions present in this repository, without a charset parameter.
var staticContentTypes = map[string]string{
	".html": "text/html",
	".js":   "text/javascript",
	".css":  "text/css",
}

// contentTypeFor returns the media type for an asset, or "" to let the
// standard library decide.
func contentTypeFor(name string) string {
	return staticContentTypes[strings.ToLower(path.Ext(name))]
}

// staticHandler serves the bundled web assets. A path that does not resolve to
// a regular file is turned away as 401, which is what the original did for a
// missing asset and for an unmatched route alike.
func staticHandler(assets fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			auth.Commence(w)
			return
		}

		name := assetPath(r.URL.Path)

		file, err := assets.Open(name)
		if err != nil {
			auth.Commence(w)
			return
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil || info.IsDir() {
			auth.Commence(w)
			return
		}

		readSeeker, ok := file.(io.ReadSeeker)
		if !ok {
			auth.Commence(w)
			return
		}

		// Set the content type explicitly: the original serves static
		// resources without a charset parameter, and Go's own extension
		// lookup would append one.
		if contentType := contentTypeFor(name); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		http.ServeContent(w, r, info.Name(), info.ModTime(), readSeeker)
	}
}
