package main

import (
	"github.com/Mpayy/e-commerce/dependency"
	"github.com/Mpayy/e-commerce/routes"
)

type Application struct {
	App    *dependency.App
	Router *routes.Router
}

func NewApplication(app *dependency.App, router *routes.Router) *Application {
	return &Application{
		App:    app,
		Router: router,
	}
}
