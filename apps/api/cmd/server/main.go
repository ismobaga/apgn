package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ismobaga/apgn/apps/api/internal/config"
	apihttp "github.com/ismobaga/apgn/apps/api/internal/http"
	"github.com/ismobaga/apgn/internal/database/postgres"
	"github.com/ismobaga/apgn/internal/orchestrator"
	rqueue "github.com/ismobaga/apgn/internal/queue/redis"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := postgres.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("connected to PostgreSQL")

	// Connect to Redis
	redisOpts, err := goredis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to parse Redis URL: %v", err)
	}
	redisClient := goredis.NewClient(redisOpts)
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("connected to Redis")

	// Create queue and orchestrator
	q := rqueue.New(redisClient)
	orch := orchestrator.New(orchestrator.Repos{
		Episodes: db,
		Jobs:     db,
	}, q)

	// Build router
	repos := apihttp.Repos{
		Shows:    db,
		Episodes: db,
		Sources:  db,
		Briefs:   db,
		Scripts:  db,
		Assets:   db,
		Jobs:     db,
	}
	router := apihttp.NewRouter(repos, orch)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("API server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
