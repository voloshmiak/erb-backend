package main

import (
	"context"
	"database/sql"
	"erb-backend/src/config"
	"erb-backend/src/controller"
	"erb-backend/src/gateway"
	"erb-backend/src/repository"
	"erb-backend/src/ticker"
	"erb-backend/src/usecase"
	"fmt"
	"log"
	"net/http"
	"time"

	"erb-backend/src/broadcaster"
	_ "erb-backend/src/docs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/swaggo/http-swagger"
)

// @title           Empty Runner Buster API
// @version         1.0.4
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
	assignmentRepository := repository.NewAssignmentRepository(conn)
	routeStepRepository := repository.NewRouteStepRepository(conn)

	b := broadcaster.New()
	matchingGateway := gateway.NewMockMatchingGateway(conn)

	listStationsUsecase := usecase.NewListStationsUseCase(stationRepository)
	fleetStatusUsecase := usecase.NewFleetStatusUseCase(wagonRepository)
	createOrderUsecase := usecase.NewCreateOrderUseCase(orderRepository, stationRepository,
		assignmentRepository, routeStepRepository, wagonRepository, b, matchingGateway)
	advanceRoutesUsecase := usecase.NewAdvanceRoutesUseCase(routeStepRepository,
		wagonRepository, assignmentRepository, orderRepository, b)
	dispatchPlannedUsecase := usecase.NewDispatchPlannedUseCase(assignmentRepository,
		routeStepRepository, wagonRepository, b)
	unloadWagonsUsecase := usecase.NewUnloadWagonsUseCase(wagonRepository, b, 30*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t := ticker.NewTicker(10*time.Second, dispatchPlannedUsecase, advanceRoutesUsecase, unloadWagonsUsecase)
	go t.Run(ctx)

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
