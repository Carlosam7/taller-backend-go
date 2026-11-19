package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Carlosam7/taller-backend-go/internal/config"
	"github.com/Carlosam7/taller-backend-go/internal/handler"
	"github.com/Carlosam7/taller-backend-go/internal/repository"
	"github.com/Carlosam7/taller-backend-go/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize database connection pool.
	ctx := context.Background()

	dbPool, err := config.NewDatabasePool(ctx)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbPool.Close()

	// Initialize dependencies.
	orderRepository := repository.NewPostgresOrderRepository(dbPool)
	orderService := service.NewOrderService(orderRepository)
	orderHandler := handler.NewOrderHandler(orderService)

	// Register routes.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/orders", orderHandler.CreateOrder)

	// Configure HTTP server.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server.
	go func() {
		log.Printf("API server listening on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Handle graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}
