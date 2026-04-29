package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client     *redis.Client
	capacity   int
	refillRate int // tokens per second
}

func NewRateLimiter(client *redis.Client, capacity, refillRate int) *RateLimiter {
	return &RateLimiter{
		client:     client,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()

		// Extract IP cleanly (important fix)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		key := "rate_limit:" + ip

		// Fetch existing data
		data, _ := rl.client.HGetAll(ctx, key).Result()

		tokens := rl.capacity
		lastRefill := time.Now().Unix()

		if len(data) > 0 {
			tokens, _ = strconv.Atoi(data["tokens"])
			lastRefill, _ = strconv.ParseInt(data["last_refill"], 10, 64)
		}

		now := time.Now().Unix()

		// Refill tokens
		elapsed := int(now - lastRefill)
		refill := elapsed * rl.refillRate

		if refill > 0 {
			tokens = min(rl.capacity, tokens+refill)
			lastRefill = now
		}

		// Block if no tokens
		if tokens <= 0 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Consume token
		tokens--

		// Save updated values
		rl.client.HSet(ctx, key, map[string]interface{}{
			"tokens":      tokens,
			"last_refill": lastRefill,
		})

		// Expire key (cleanup)
		rl.client.Expire(ctx, key, time.Minute*2)

		next.ServeHTTP(w, r)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}