package main

import (
	"database/sql"
	"erb-backend/src/config"
	"erb-backend/src/controller"
	"erb-backend/src/repository"
	"erb-backend/src/usecase"
	"fmt"
	"log"
	"net/http"

	"erb-backend/src/broadcaster"
	_ "erb-backend/src/docs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/swaggo/http-swagger"
)

// @title           Empty Runner Buster API
// @version         1.0.3
// @description		This is the API documentation for the Empty Runner Buster application
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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require",
		cfg.Database.Host, cfg.Database.User,
		cfg.Database.UserPassword, cfg.Database.Name)

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	if err = conn.Ping(); err != nil {
		return err
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	router := http.NewServeMux()

	stationRepository := repository.NewStationRepository(conn)
	wagonRepository := repository.NewWagonRepository(conn)
	orderRepository := repository.NewOrderRepository(conn)

	b := broadcaster.New()

	listStationsUsecase := usecase.NewListStationsUseCase(stationRepository)
	fleetStatusUsecase := usecase.NewFleetStatusUseCase(wagonRepository)
	createOrderUsecase := usecase.NewCreateOrderUseCase(orderRepository, b)

	healthController := controller.NewHealthController()
	docsController := controller.NewDocsController()
	eventsStreamController := controller.NewEventsStreamController(b)
	listStationsController := controller.NewListStationsController(listStationsUsecase)
	fleetStatusController := controller.NewFleetStatusController(fleetStatusUsecase)
	createOrderController := controller.NewCreateOrderController(createOrderUsecase)

	router.Handle("GET /api/health", healthController)
	router.Handle("GET /api/docs", docsController)
	router.Handle("GET /api/stations", listStationsController)
	router.Handle("GET /api/fleet/status", fleetStatusController)
	router.Handle("POST /api/orders", createOrderController)
	router.Handle("GET /api/events/stream", eventsStreamController)
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
