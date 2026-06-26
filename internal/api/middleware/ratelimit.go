package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/your-org/isp-billing/internal/pkg/logger"
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
		// Get client IP
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}

		key := "ratelimit:" + ip
		ctx := context.Background()

		// Get current count
		count, err := rl.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			logger.Error("Rate limit check failed", map[string]interface{}{"error": err})
			next.ServeHTTP(w, r)
			return
		}

		if count >= rl.limit {
			respondError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		// Increment counter
		pipe := rl.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rl.window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			logger.Error("Rate limit increment failed", map[string]interface{}{"error": err})
		}

		next.ServeHTTP(w, r)
	})
}
