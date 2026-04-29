package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"url-shortener/internal/models"
	"url-shortener/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrURLNotFound = errors.New("url not found or expired")
)

type URLService struct {
	redisClient    *redis.Client
	urlsCollection *mongo.Collection
}

func NewURLService(redisClient *redis.Client, urlsCollection *mongo.Collection) *URLService {
	return &URLService{
		redisClient:    redisClient,
		urlsCollection: urlsCollection,
	}
}

// ShortenURL generates a short URL for the given long URL
func (s *URLService) ShortenURL(ctx context.Context, longURL string, ttlSeconds int) (string, error) {
	// 1. Get auto-incrementing ID
	id, err := s.redisClient.Incr(ctx, "global:next_id").Result()
	if err != nil {
		return "", fmt.Errorf("failed to generate unique ID: %w", err)
	}

	// 2. Encode to Base62 (8-character non-sequential combination)
	shortCode := utils.GenerateHashID(uint64(id))

	// 3. Store the mapping in MongoDB
	var expiresAt *time.Time
	if ttlSeconds > 0 {
		exp := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
		expiresAt = &exp
	}

	doc := models.URLDocument{
		ShortCode:   shortCode,
		OriginalURL: longURL,
		Clicks:      0,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}

	_, err = s.urlsCollection.InsertOne(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("failed to save URL mapping to mongodb: %w", err)
	}

	// 4. Cache in Redis
	urlKey := fmt.Sprintf("url:%s", shortCode)
	expiration := time.Duration(ttlSeconds) * time.Second

	err = s.redisClient.Set(ctx, urlKey, longURL, expiration).Err()
	if err != nil {
		// Log the error but don't fail the request
		fmt.Printf("failed to cache URL mapping: %v\n", err)
	}

	return shortCode, nil
}

// GetOriginalURL retrieves the original URL and increments the click count
func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	urlKey := fmt.Sprintf("url:%s", shortCode)

	// 1. Try to fetch from Redis cache
	longURL, err := s.redisClient.Get(ctx, urlKey).Result()
	if err == nil {
		// Cache hit
		s.incrementClicksAsync(shortCode)
		return longURL, nil
	} else if !errors.Is(err, redis.Nil) {
		// Log redis error but fallback to mongodb
		fmt.Printf("redis cache error: %v\n", err)
	}

	// 2. Cache miss, fetch from MongoDB
	var doc models.URLDocument
	err = s.urlsCollection.FindOne(ctx, bson.M{"short_code": shortCode}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrURLNotFound
		}
		return "", fmt.Errorf("failed to retrieve URL from mongodb: %w", err)
	}

	// 3. Check if expired
	if doc.ExpiresAt != nil && doc.ExpiresAt.Before(time.Now()) {
		return "", ErrURLNotFound
	}

	// 4. Update Redis cache for subsequent requests
	var expiration time.Duration
	if doc.ExpiresAt != nil {
		expiration = time.Until(*doc.ExpiresAt)
		if expiration < 0 {
			expiration = 0 // Just in case
		}
	} else {
		expiration = 0
	}

	err = s.redisClient.Set(ctx, urlKey, doc.OriginalURL, expiration).Err()
	if err != nil {
		fmt.Printf("failed to update cache: %v\n", err)
	}

	// 5. Increment click count asynchronously
	s.incrementClicksAsync(shortCode)

	return doc.OriginalURL, nil
}

// incrementClicksAsync increments the click counter in MongoDB asynchronously
func (s *URLService) incrementClicksAsync(shortCode string) {
	go func() {
		// Use background context since request context may be canceled
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := s.urlsCollection.UpdateOne(
			ctx,
			bson.M{"short_code": shortCode},
			bson.M{"$inc": bson.M{"clicks": 1}},
		)
		if err != nil {
			fmt.Printf("failed to increment click counter for %s in mongodb: %v\n", shortCode, err)
		}
	}()
}
