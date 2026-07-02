package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/t0fox/subscription-aggregator-api/docs"
	"github.com/t0fox/subscription-aggregator-api/internal/config"
	"github.com/t0fox/subscription-aggregator-api/internal/handlers"
	"github.com/t0fox/subscription-aggregator-api/internal/middleware"
	"github.com/t0fox/subscription-aggregator-api/internal/repository"
	"github.com/t0fox/subscription-aggregator-api/internal/service"
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

	r := gin.Default()
	r.Use(middleware.Logging())

	connStr := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=prefer"
	conn, err := connectWithRetry(context.Background(), connStr, 30, time.Second)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(context.Background())

	subscriptionRepo := repository.NewSubscriptionRepository(conn)
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

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func connectWithRetry(ctx context.Context, connStr string, attempts int, delay time.Duration) (*pgx.Conn, error) {
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		conn, err := pgx.Connect(ctx, connStr)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		log.Printf("Database is not ready yet, retrying (%d/%d): %v", attempt, attempts, err)
		time.Sleep(delay)
	}

	return nil, lastErr
}
