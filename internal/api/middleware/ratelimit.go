package middleware

import (
	//"context"
	"log"
	"net/http"
	"strings" // Added missing import for string parsing
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redis.Client
	limit       int
	window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP safely
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}

		key := "ratelimit:" + ip
		ctx := r.Context() // Use the request's context rather than a detached Background context

		// Get current count
		count, err := rl.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			log.Printf("[RATELIMIT ERROR] Rate limit read check failed for IP %s: %v", ip, err)
			next.ServeHTTP(w, r)
			return
		}

		if count >= rl.limit {
			respondError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			return
		}

		// Increment counter atomically within the evaluation window
		pipe := rl.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rl.window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Printf("[RATELIMIT ERROR] Redis transaction execution pipeline failed for IP %s: %v", ip, err)
		}

		next.ServeHTTP(w, r)
	})
}
