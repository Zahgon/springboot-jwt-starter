// Command springboot-jwt-starter serves a stateless JWT authentication API and
// the single-page UI that drives it.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
)

// addr is the address the server listens on.
const addr = ":8080"

//go:embed all:web
var webFS embed.FS

func main() {
	assets, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	application, err := app.New(assets, app.Options{})
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}
	defer application.Close()

	log.Printf("Started Application on port %s", addr)
	if err := http.ListenAndServe(addr, application.Handler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
