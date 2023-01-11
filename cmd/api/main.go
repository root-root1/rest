package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/root-root1/rest/internal/data"
	"github.com/root-root1/rest/internal/jsonlog"
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

	Limiter struct {
		rps    float64
		bust   int
		enable bool
	}
}

type Application struct {
	Config  Config
	Logger  *jsonlog.Logger
	Models  data.Models
	Version string
}

func main() {
	err := godotenv.Load()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	if err != nil {
		logger.PrintError(err, nil)
	}

	var cfg Config

	flag.IntVar(&cfg.Port, "Port", 8000, "Port Variable by Default 8000")
	flag.StringVar(&cfg.Env, "Environment Variable", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("REST_DB_DSN"), "Postgres DSN (Data Source Name)")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 15, "PostgreSQL max open Connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 15, "PostgreSQL max idle Connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max Connection Time")
	flag.Float64Var(&cfg.Limiter.rps, "limiter-rps", 2, "Rate Limiter Maximum Request per second")
	flag.IntVar(&cfg.Limiter.bust, "limiter-bust", 4, "Rate Limiter Maximum Bust")
	flag.BoolVar(&cfg.Limiter.enable, "enable", true, "Enable Rate Limiter")

	flag.Parse()

	db, err := openDb(cfg)

	app := &Application{
		Config:  cfg,
		Logger:  logger,
		Models:  data.NewModel(db),
		Version: version,
	}

	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.PrintError(err, nil)
		}
	}(db)

	logger.PrintInfo("Connection to Database has Done", nil)

	err = app.serve()

	if err != nil {
		logger.PrintFatal(err, nil)
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
