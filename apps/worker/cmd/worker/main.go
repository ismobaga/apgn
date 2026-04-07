package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ismobaga/apgn/apps/worker/internal/jobs"
	"github.com/ismobaga/apgn/internal/database/postgres"
	"github.com/ismobaga/apgn/internal/orchestrator"
	rqueue "github.com/ismobaga/apgn/internal/queue/redis"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/podgen?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")

	db, err := postgres.New(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("connected to PostgreSQL")

	redisOpts, err := goredis.ParseURL(redisURL)
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

	q := rqueue.New(redisClient)
	orch := orchestrator.New(orchestrator.Repos{
		Episodes: db,
		Jobs:     db,
	}, q)

	repos := jobs.Repos{
		Shows:    db,
		Episodes: db,
		Sources:  db,
		Briefs:   db,
		Scripts:  db,
		Assets:   db,
		Jobs:     db,
	}

	// Providers are optional; nil means the stage is skipped
	dispatcher := jobs.NewDispatcher(repos, orch, q, nil, nil, nil)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("shutting down worker...")
		cancel()
	}()

	dispatcher.Run(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
