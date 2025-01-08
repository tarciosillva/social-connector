package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"social-connector/internal/config"
	"social-connector/internal/infra/handlers"
	"social-connector/internal/infra/logger"
	"social-connector/internal/infra/routes"
	"social-connector/internal/middleware"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	config.LoadEnv()

	ctx := context.Background()
	log := logger.NewLogger(ctx, true)

	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(log))

	verifyToken := config.GetEnv("API_KEY")

	transactionHandlers := handlers.NewHttpHandlers(log, verifyToken)

	routes := routes.NewRoutes(
		router,
		transactionHandlers,
	)

	routes.Init()

	port := config.GetEnv("PORT")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Info(fmt.Sprintf("Server is running on port %s", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(fmt.Sprintf("Error running HTTP server: %s", err))
			os.Exit(1)
		}
	}()

	<-stop
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	} else {
		log.Info("Server stopped gracefully.")
	}
}
