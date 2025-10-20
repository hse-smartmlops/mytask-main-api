package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"emplacc-api/api/v1/dto"
	"emplacc-api/internal/config"

	"github.com/labstack/echo/v4"
)

type rateLimiter struct {
	config config.RateLimitConfig
	mu     sync.Mutex
	store  map[string]*limitBucket
}

type limitBucket struct {
	count int
	reset time.Time
}

func NewRateLimiter(cfg config.RateLimitConfig) echo.MiddlewareFunc {
	if cfg.User.Requests <= 0 && cfg.IP.Requests <= 0 {
		return nil
	}

	limiter := &rateLimiter{
		config: cfg,
		store:  make(map[string]*limitBucket),
	}

	return limiter.handle
}

func (r *rateLimiter) handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		now := time.Now()

		if r.config.User.Requests > 0 {
			if userKey := userRateKey(c); userKey != "" {
				if ok, retryAfter := r.allow(now, "user:"+userKey, r.config.User); !ok {
					return limitExceeded(c, retryAfter)
				}
			}
		}

		if r.config.IP.Requests > 0 {
			if ip := c.RealIP(); ip != "" {
				if ok, retryAfter := r.allow(now, "ip:"+ip, r.config.IP); !ok {
					return limitExceeded(c, retryAfter)
				}
			}
		}

		return next(c)
	}
}

func (r *rateLimiter) allow(now time.Time, key string, rule config.RateLimitRule) (bool, time.Duration) {
	window := rule.Window
	if window <= 0 {
		window = time.Minute
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, ok := r.store[key]
	if !ok || now.After(bucket.reset) {
		bucket = &limitBucket{count: 0, reset: now.Add(window)}
		r.store[key] = bucket
	}

	if bucket.count >= rule.Requests {
		return false, bucket.reset.Sub(now)
	}

	bucket.count++
	return true, bucket.reset.Sub(now)
}

func userRateKey(c echo.Context) string {
	if raw := c.Get("auth_token"); raw != nil {
		if token, ok := raw.(string); ok && token != "" {
			return token
		}
	}
	return ""
}

func limitExceeded(c echo.Context, retryAfter time.Duration) error {
	resp := dto.NewError("rate_limited", "rate limit exceeded")
	if traceID := c.Response().Header().Get(echo.HeaderXRequestID); traceID != "" {
		resp.Meta.TraceID = traceID
	} else {
		resp.Meta.TraceID = c.Request().Header.Get(echo.HeaderXRequestID)
	}

	if retryAfter > 0 {
		seconds := int(retryAfter.Round(time.Second).Seconds())
		if seconds < 1 {
			seconds = 1
		}
		c.Response().Header().Set(echo.HeaderRetryAfter, strconv.Itoa(seconds))
	}

	return c.JSON(http.StatusTooManyRequests, resp)
}
