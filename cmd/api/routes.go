package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *Application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	router.HandlerFunc(http.MethodGet, "/api/v1/health-check", app.healthCheckHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/movie", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/movie/:id", app.getMovieById)
	router.HandlerFunc(http.MethodPatch, "/api/v1/movie/:id", app.UpdateMovie)
	router.HandlerFunc(http.MethodDelete, "/api/v1/delete/:id", app.deleteMovie)

	return router
}
