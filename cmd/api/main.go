package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	//Declare an instance of the config struct
	var cfg config

	// Read the value of the port and env command-line flags into the config struct. We default to using the port number 4000 and the environment "development" if no corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", 6969, "API Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development | staging | production)")

	//export DSN="postgres://admin:password@localhost:5432/apimovie?sslmode=disable" in shell
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("DSN"), "postgres")
	flag.IntVar(&cfg.db.maxOpenConns, "DB-maxOpenConnections", 25, "postgresSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "DB-maxIdleConnections", 25, "postgresSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "DB-max-idle-time", "15m", "postgreSQL max connection idle time")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Printf("database connection pool established")

	migrateDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatalf("failed to create Postgres migration driver: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", migrateDriver)
	if err != nil {
		logger.Fatalf("Failed to initialize migrator: %v", err)
	}

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Fatalf("Failed to apply migrations: %v", err)
	}

	app := &application{
		config: cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.healthCheckHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func (app *application) faliedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func openDB(cfg config) (*sql.DB, error) {
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
