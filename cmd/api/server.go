package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/root-root1/rest/internal/jsonlog"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *Application) serve() error {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	shutdown := make(chan error)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.Port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ErrorLog:          log.New(logger, "", 0),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		s := <-quit
		app.Logger.PrintInfo("Caught Signal", map[string]string{
			"signal": s.String(),
		})

		os.Exit(0)
	}()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		app.Logger.PrintInfo("Shutting Down Server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdown <- srv.Shutdown(ctx)
	}()

	logger.PrintInfo("Starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.Config.Env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}

	app.Logger.PrintInfo("Stopping Server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
