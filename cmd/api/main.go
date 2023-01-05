package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/root-root1/rest/internal/data"
	"log"
	"net/http"
	"os"
	"time"
)

var version = "1.0.0"

type Config struct {
	Port int
	Env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type Application struct {
	Config     Config
	InfoLogger *log.Logger
	ErrorLog   *log.Logger
	Models     data.Models
	Version    string
}

func (app *Application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.Port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.InfoLogger.Println(fmt.Sprintf("Starting HTTP Server in %s Mode on Port %d", app.Config.Env, app.Config.Port))
	return srv.ListenAndServe()
}

func main() {
	err := godotenv.Load()

	infoLog := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	if err != nil {
		errorLog.Println("Failed to load .env file")
	}

	var cfg Config

	flag.IntVar(&cfg.Port, "Port", 8000, "Port Variable by Default 8000")
	flag.StringVar(&cfg.Env, "Environment Variable", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("REST_DB_DSN"), "Postgres DSN (Data Source Name)")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 15, "PostgreSQL max open Connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 15, "PostgreSQL max idle Connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max Connection Time")

	flag.Parse()

	db, err := openDb(cfg)
	
	app := &Application{
		Config:     cfg,
		InfoLogger: infoLog,
		ErrorLog:   errorLog,
		Models:     data.NewModel(db),
		Version:    version,
	}

	if err != nil {
		app.ErrorLog.Fatal(err)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			app.ErrorLog.Println("Failed To Closing the Database Connection")
		}
	}(db)

	app.InfoLogger.Println("Connection to Database has Done")

	err = app.serve()

	if err != nil {
		app.ErrorLog.Println(err)
		log.Fatal(err)
	}
}

func openDb(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
