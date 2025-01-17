package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"social-connector/internal/config"
	"social-connector/internal/domain/entities"
	Iservices "social-connector/internal/domain/interfaces/services"
	"social-connector/internal/infra/handlers"
	"social-connector/internal/infra/logger"
	"social-connector/internal/infra/repository"
	"social-connector/internal/infra/routes"
	"social-connector/internal/infra/services"
	"social-connector/internal/middleware"
	client "social-connector/internal/pkg"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	config.LoadEnv()

	ctx := context.Background()
	log := logger.NewLogger(ctx, true)

	mongoClient := client.MongoClient()
	userContextDB := mongoClient.Database("UserContext")

	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(log))

	userContextRepo := repository.NewMongoRepository[entities.UserContext](userContextDB)

	var userContextSvc Iservices.IUserContextService = services.NewUserContextService(userContextRepo, ctx, log)
	var queryAIService Iservices.IQueryAIService = services.NewQueryAIService(log)

	httpClient := http.Client{}

	verifyToken := config.GetEnv("API_KEY")

	transactionHandlers := handlers.NewHttpHandlers(log, verifyToken, userContextSvc, queryAIService)
	infobipHandlers := handlers.NewInfobipHandlers(log, userContextSvc, queryAIService, &httpClient)

	routes := routes.NewRoutes(
		router,
		transactionHandlers,
		infobipHandlers,
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
