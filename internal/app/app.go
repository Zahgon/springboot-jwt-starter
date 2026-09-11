// Package app wires the application's components together, which is what the
// Spring context did.
package app

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/bfwg/springboot-jwt-starter/internal/clock"
	"github.com/bfwg/springboot-jwt-starter/internal/config"
	"github.com/bfwg/springboot-jwt-starter/internal/db"
	"github.com/bfwg/springboot-jwt-starter/internal/repository"
	"github.com/bfwg/springboot-jwt-starter/internal/rest"
	"github.com/bfwg/springboot-jwt-starter/internal/security"
	"github.com/bfwg/springboot-jwt-starter/internal/security/auth"
	"github.com/bfwg/springboot-jwt-starter/internal/security/crypto"
	"github.com/bfwg/springboot-jwt-starter/internal/server"
	"github.com/bfwg/springboot-jwt-starter/internal/service"
)

// App is a wired application: its HTTP handler plus the components tests need
// to reach in order to pin the clock or inspect the token helper.
type App struct {
	Config      *config.Config
	DB          *sql.DB
	Handler     http.Handler
	TokenHelper *security.TokenHelper
	Users       service.UserService
	UserDetails service.UserDetailsService
}

// Options adjust how an application is wired. The zero value is production.
type Options struct {
	// Clock overrides the time source. Defaults to the wall clock.
	Clock clock.Provider
	// DBName names the in-memory database, so tests can hold independent ones.
	DBName string
	// UserDetails overrides principal lookup. Defaults to the repository-backed
	// service.
	UserDetails service.UserDetailsService
}

// New wires an application from the embedded configuration and assets.
func New(assets fs.FS, opts Options) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	timeProvider := opts.Clock
	if timeProvider == nil {
		timeProvider = clock.System{}
	}
	name := opts.DBName
	if name == "" {
		name = "bfwg"
	}

	database, err := db.Open(name)
	if err != nil {
		return nil, err
	}

	users := repository.New(database)
	encoder := crypto.NewBCryptPasswordEncoder()

	var userDetails service.UserDetailsService = service.NewCustomUserDetailsService(users)
	if opts.UserDetails != nil {
		userDetails = opts.UserDetails
	}

	tokenHelper := security.NewTokenHelper(
		cfg.App.Name, cfg.JWT.Secret, cfg.JWT.ExpiresIn, cfg.JWT.Header, timeProvider)

	userService := service.NewUserService(users, encoder, userDetails, timeProvider)

	handler := server.New(
		assets,
		rest.NewAuthenticationController(tokenHelper, userService),
		rest.NewUserController(userService),
		auth.TokenAuthenticationFilter(tokenHelper, userDetails),
	)

	return &App{
		Config:      cfg,
		DB:          database,
		Handler:     handler,
		TokenHelper: tokenHelper,
		Users:       userService,
		UserDetails: userDetails,
	}, nil
}

// Close releases the application's database.
func (a *App) Close() error {
	if a.DB == nil {
		return nil
	}
	if err := a.DB.Close(); err != nil {
		return fmt.Errorf("app: closing database: %w", err)
	}
	return nil
}
