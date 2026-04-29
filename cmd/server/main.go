package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/handlers"
	"url-shortener/internal/services"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()

	// Initialize Redis Client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")

	// Initialize MongoDB Client
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test MongoDB connection
	ctxMongo, cancelMongo := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelMongo()
	if err := mongoClient.Ping(ctxMongo, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB successfully!")

	mongoDb := mongoClient.Database(cfg.MongoDBName)
	urlsCollection := mongoDb.Collection("urls")

	// Initialize Service
	urlService := services.NewURLService(redisClient, urlsCollection)

	// Initialize Handler
	// You can pass the actual domain from config in a real app
	domain := fmt.Sprintf("http://localhost:%s", cfg.ServerPort)
	urlHandler := handlers.NewURLHandler(urlService, domain)

	// Setup Router
	r := mux.NewRouter()

	// Define Routes
	r.HandleFunc("/shorten", urlHandler.Shorten).Methods("POST")
	r.HandleFunc("/{short_code}", urlHandler.Redirect).Methods("GET")

	// Start Server
	addr := ":" + cfg.ServerPort
	log.Printf("Server is starting on port %s...", cfg.ServerPort)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
