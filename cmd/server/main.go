package main

import (
	"CourseJob/internal/config"
	"CourseJob/internal/http"
	"CourseJob/internal/storage/postgres"
	"context"
	"log"
	nethttp "net/http"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	DB, err := postgres.NewPool(ctx, cfg.DataBAseURL)
	if err != nil {
		log.Fatalf("failed connected to database %v", err)
	}
	defer DB.Close()

	handler := &http.Handler{DB.Pool}

	router := http.NewRouter(handler)

	server := &nethttp.Server{
		Handler: router,
		Addr:    cfg.HTTPAddr,
	}
	log.Printf("server started on: %s", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
		log.Fatalf("http server start failed: %v", err)
	}

}
