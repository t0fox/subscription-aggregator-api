package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/t0fox/subscription-aggregator-api/docs"
	"github.com/t0fox/subscription-aggregator-api/internal/config"
	"github.com/t0fox/subscription-aggregator-api/internal/handlers"
	"github.com/t0fox/subscription-aggregator-api/internal/middleware"
	"github.com/t0fox/subscription-aggregator-api/internal/repository"
	"github.com/t0fox/subscription-aggregator-api/internal/service"
	"github.com/t0fox/subscription-aggregator-api/pkg/database"
)

// @title Subscription Service API
// @version 1.0
// @description API for managing online subscription records.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Context cancelled on Ctrl+C (SIGINT) or docker stop (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r := gin.Default()
	r.Use(middleware.Logging())

	db, err := database.NewDatabase(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	subscriptionRepo := repository.NewSubscriptionRepository(db.Pool)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)

	api := r.Group("/api/v1")
	{
		api.POST("/subscriptions", subscriptionHandler.Create)
		api.GET("/subscriptions", subscriptionHandler.GetAll)
		api.GET("/subscriptions/:id", subscriptionHandler.GetByID)
		api.PUT("/subscriptions/:id", subscriptionHandler.Update)
		api.DELETE("/subscriptions/:id", subscriptionHandler.Delete)
		api.POST("/subscriptions/sum", subscriptionHandler.GetSum)
	}

	url := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Run the server in a goroutine so it doesn't block signal handling.
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Block until a shutdown signal arrives.
	<-ctx.Done()
	log.Println("Shutdown signal received, stopping server gracefully...")

	// Give in-flight requests up to 10 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed, forcing exit: %v", err)
	} else {
		log.Println("Server stopped cleanly")
	}
}
