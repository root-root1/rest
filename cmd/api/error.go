package main

import (
	"fmt"
	"net/http"
)

func (app *Application) LogError(r *http.Request, err error) {
	app.Logger.PrintError(err, map[string]string{
		"request-method": r.Method,
		"request-url":    r.URL.String(),
	})
}

func (app *Application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.LogError(r, err)
		w.WriteHeader(500)
	}
}

func (app *Application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.LogError(r, err)
	message := "Server Encountered a problem and couldn't process you Request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The Requested Resource Could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *Application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("this %s Method is not Allowed for the Resouce", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *Application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *Application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "Unable to Edit the Record Due to Edit Conflict. please try Again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "Rate Limit Exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}
