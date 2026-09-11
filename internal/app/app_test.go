package app_test

import (
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/app"
	"github.com/bfwg/springboot-jwt-starter/internal/apptest"
)

func TestContextLoads(t *testing.T) {
	application := apptest.New(t, app.Options{})

	if application.Handler == nil {
		t.Error("the application has no HTTP handler")
	}
	if application.TokenHelper == nil {
		t.Error("the application has no token helper")
	}
	if application.Config.App.Name != "springboot-jwt-demo" {
		t.Errorf("app.name = %q, want %q", application.Config.App.Name, "springboot-jwt-demo")
	}
}
