package main

import (
	"net/http"
)

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status":      "Available",
		"Environment": app.Config.Env,
		"Version":     app.Version,
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"Movie": data}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
