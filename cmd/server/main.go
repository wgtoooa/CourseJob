package main

import (
	"CourseJob/internal/config"
	"CourseJob/internal/service"
	"CourseJob/internal/storage/postgres"
	http2 "CourseJob/internal/transport/http"
	"context"
	"log"
	nethttp "net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	log.Println("config loaded")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	tx := postgres.NewTxManager(db.Pool)
	attendanceService := service.NewAttendanceService(tx)

	handler := http2.NewHandler(db.Pool, attendanceService)
	router := http2.NewRouter(handler)

	server := &nethttp.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		log.Printf("server started on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			log.Fatalf("http server start failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}
