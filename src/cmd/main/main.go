package main

import (
	"database/sql"
	"erb-backend/src/config"
	"erb-backend/src/controller"
	"fmt"
	"log"
	"net/http"

	_ "erb-backend/src/docs"

	"cloud.google.com/go/cloudsqlconn/postgres/pgxv5"
	"github.com/swaggo/http-swagger"
)

const (
	cloudDriverName = "cloudsql-postgres"
)

// @title           Empty Runner Buster API
// @version         1.0
// @description		This is the API documentation for the Empty Runner Buster application
// @host            localhost:8080
// @BasePath        /api
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	cleanup, err := pgxv5.RegisterDriver(cloudDriverName)
	if err != nil {
		return err
	}
	defer func() {
		err = cleanup()
		if err != nil {
			log.Println("Error during driver cleanup:", err)
		}
	}()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.ConnectionName, cfg.Database.User,
		cfg.Database.UserPassword, cfg.Database.Name)

	conn, err := sql.Open(cloudDriverName, dsn)
	if err != nil {
		return err
	}

	if err = conn.Ping(); err != nil {
		return err
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	router := http.NewServeMux()

	healthController := controller.NewHealthController()
	docsController := controller.NewDocsController()

	router.Handle("/api/health", healthController)
	router.Handle("/api/docs", docsController)
	router.Handle("/swagger/", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Println("Starting server on port", cfg.Port)

	if err = server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
