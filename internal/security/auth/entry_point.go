package auth

import "net/http"

// Commence turns away a request that is not allowed to proceed.
//
// This is the single failure surface of the application: every authentication
// and every authorization failure — and every unmatched route, which the
// original funnelled here too — answers 401 with a completely empty body. There
// is no error envelope and no 403.
func Commence(w http.ResponseWriter) {
	// Spring's sendError path replaces any content type already negotiated.
	w.Header().Del("Content-Type")
	w.WriteHeader(http.StatusUnauthorized)
}
