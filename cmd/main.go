package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	delivery "github.com/opusdvs/DonWeather-ms-georesolve/internal/delivery/http"
	"github.com/opusdvs/DonWeather-ms-georesolve/internal/delivery/middleware"
	"github.com/opusdvs/DonWeather-ms-georesolve/internal/repository"
	"github.com/opusdvs/DonWeather-ms-georesolve/internal/usecase"
)

func main() {
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatal("DB_USER environment variable is required")
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Fatal("DB_HOST environment variable is required")
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Fatal("DB_PORT environment variable is required")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME environment variable is required")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresRepoCity(db)
	service := usecase.NewCityService(repo)
	handler := delivery.NewCityHandler(service)

	healthHandler := delivery.NewHealthHandler(db)

	mainMux := http.NewServeMux()
	mainMux.HandleFunc("/health/liveness", healthHandler.LivenessProbe)
	mainMux.HandleFunc("/health/readiness", healthHandler.ReadinessProbe)
	mainMux.Handle("/api/v1/georesolve", middleware.CORSMiddleware(middleware.TraceMiddleware(http.HandlerFunc(handler.GetCityByCoordinates))))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mainMux,
	}

	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}(server)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}

	log.Println("Server stopped")
}
