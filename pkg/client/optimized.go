// Package client provides an optimized HTTP client for API interactions.
// It includes connection pooling, retry logic with exponential backoff,
// and rate limiting capabilities.
package client

import (
	"context"
	"fmt"
	"img-cli/pkg/errors"
	"img-cli/pkg/logger"
	"math"
	"net/http"
	"sync"
	"time"
)

// RetryConfig defines retry behavior for API requests
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// DefaultRetryConfig returns sensible retry defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens    int
	maxTokens int
	interval  time.Duration
	mu        sync.Mutex
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10 // default to 10 RPS
	}

	interval := time.Duration(float64(time.Second) / requestsPerSecond)

	return &RateLimiter{
		tokens:     1,
		maxTokens:  int(requestsPerSecond),
		interval:   interval,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.interval)

	if tokensToAdd > 0 {
		r.tokens = min(r.maxTokens, r.tokens+tokensToAdd)
		r.lastRefill = now
	}

	// If no tokens available, wait
	if r.tokens <= 0 {
		waitTime := r.interval
		logger.Debug("Rate limit reached, waiting", "wait_time", waitTime)

		select {
		case <-time.After(waitTime):
			r.tokens = 1
			r.lastRefill = time.Now()
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	r.tokens--
	return nil
}

// OptimizedClient wraps an HTTP client with performance optimizations
type OptimizedClient struct {
	httpClient  *http.Client
	retryConfig *RetryConfig
	rateLimiter *RateLimiter
	baseURL     string
	apiKey      string
}

// Config holds configuration for the optimized client
type Config struct {
	BaseURL           string
	APIKey            string
	Timeout           time.Duration
	MaxIdleConns      int
	MaxConnsPerHost   int
	IdleConnTimeout   time.Duration
	RequestsPerSecond float64
	RetryConfig       *RetryConfig
}

// DefaultConfig returns default client configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout:           180 * time.Second,
		MaxIdleConns:      100,
		MaxConnsPerHost:   10,
		IdleConnTimeout:   90 * time.Second,
		RequestsPerSecond: 2.0, // Conservative default
		RetryConfig:       DefaultRetryConfig(),
	}
}

// NewOptimizedClient creates a new optimized HTTP client
func NewOptimizedClient(config *Config) *OptimizedClient {
	if config == nil {
		config = DefaultConfig()
	}

	// Create transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &OptimizedClient{
		httpClient:  httpClient,
		retryConfig: config.RetryConfig,
		rateLimiter: NewRateLimiter(config.RequestsPerSecond),
		baseURL:     config.BaseURL,
		apiKey:      config.APIKey,
	}
}

// DoWithRetry executes an HTTP request with retry logic
func (c *OptimizedClient) DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	backoff := c.retryConfig.InitialBackoff

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Apply rate limiting
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, errors.Wrap(err, errors.APIError, "rate limiter cancelled")
		}

		// Clone request for retry
		reqClone := req.Clone(ctx)

		// Execute request
		resp, err := c.httpClient.Do(reqClone)

		// Success or non-retryable error
		if err == nil && resp.StatusCode < 500 {
			if resp.StatusCode == 429 {
				// Rate limit hit, back off more aggressively
				logger.Warn("API rate limit hit", "attempt", attempt+1)
				backoff = c.retryConfig.MaxBackoff
			} else {
				// Success or client error (4xx)
				return resp, nil
			}
		}

		if err != nil {
			lastErr = err
			logger.Warn("Request failed",
				"attempt", attempt+1,
				"max_attempts", c.retryConfig.MaxRetries+1,
				"error", err)
		} else {
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			resp.Body.Close()
			logger.Warn("Server error",
				"attempt", attempt+1,
				"status", resp.StatusCode)
		}

		// Don't retry if this was the last attempt
		if attempt == c.retryConfig.MaxRetries {
			break
		}

		// Wait with exponential backoff
		select {
		case <-time.After(backoff):
			backoff = time.Duration(math.Min(
				float64(backoff)*c.retryConfig.BackoffFactor,
				float64(c.retryConfig.MaxBackoff),
			))
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, errors.Wrapf(lastErr, errors.APIError, "request failed after %d attempts", c.retryConfig.MaxRetries+1)
}

// Close cleans up client resources
func (c *OptimizedClient) Close() {
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}