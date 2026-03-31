package main

import (
	"erb-backend/src/config"
	"erb-backend/src/controller"
	"log"
	"net/http"
)

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

	router := http.NewServeMux()

	healthController := controller.NewHealthController()

	router.Handle("/health", healthController)

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
