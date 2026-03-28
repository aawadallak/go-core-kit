// Package eventhttp provides eventhttp functionality.
package eventhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aawadallak/go-core-kit/core/event"

	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type RetryConfig struct {
	MaxRetries     int                  `json:"max_retries"`
	InitialBackoff time.Duration        `json:"initial_backoff"`
	MaxBackoff     time.Duration        `json:"max_backoff"`
	BackoffFactor  float64              `json:"backoff_factor"`
	JitterEnabled  bool                 `json:"jitter_enabled"`
	RequestTimeout time.Duration        `json:"request_timeout"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker"`
}

type CircuitBreakerConfig struct {
	Enabled          bool          `json:"enabled"`
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
}

type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

type CircuitBreaker struct {
	config       CircuitBreakerConfig
	state        CircuitBreakerState
	failures     int
	successes    int
	lastFailTime time.Time
	mutex        sync.RWMutex
}

type WebhookRequest struct {
	EventName     string          `json:"event_name"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

// HTTPEventDispatcher with advanced retry, backoff and circuit breaker
type HTTPEventDispatcher struct {
	client         *resty.Client
	serviceURL     string
	config         RetryConfig
	circuitBreaker *CircuitBreaker
}

// HTTPEventDispatcherOption allows configuring the dispatcher
type HTTPEventDispatcherOption func(*HTTPEventDispatcher)

// WithRetryConfig sets the retry configuration
func WithRetryConfig(config RetryConfig) HTTPEventDispatcherOption { //nolint:gocritic // hugeParam
	return func(d *HTTPEventDispatcher) {
		d.config = config
	}
}

// WithServiceURL sets the base URL for the webhook service
func WithServiceURL(url string) HTTPEventDispatcherOption {
	return func(d *HTTPEventDispatcher) {
		d.serviceURL = url
	}
}

func WithClient(client *resty.Client) HTTPEventDispatcherOption {
	return func(d *HTTPEventDispatcher) {
		d.client = client
	}
}

// NewEventDispatcher creates a new dispatcher with configurable options
func NewEventDispatcher(opts ...HTTPEventDispatcherOption) *HTTPEventDispatcher {
	client := resty.New()

	dispatcher := &HTTPEventDispatcher{
		client:     client,
		serviceURL: "http://localhost:8080",
		config: RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
			JitterEnabled:  true,
			RequestTimeout: 10 * time.Second,
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:          true,
				FailureThreshold: 5,
				SuccessThreshold: 3,
				Timeout:          60 * time.Second,
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(dispatcher)
	}

	// Initialize circuit breaker if enabled
	if dispatcher.config.CircuitBreaker.Enabled {
		dispatcher.circuitBreaker = &CircuitBreaker{
			config: dispatcher.config.CircuitBreaker,
			state:  CircuitClosed,
		}
	}

	return dispatcher
}

// Dispatch implements event.Publisher with robust retry mechanism.
func (h *HTTPEventDispatcher) Dispatch(ctx context.Context, evt *event.Record) error {
	// Check circuit breaker
	if h.circuitBreaker != nil && !h.circuitBreaker.canExecute() {
		logger.Of(ctx).WarnS("HTTPEventDispatcher:CircuitBreakerOpen",
			logger.WithValue("event_id", evt.ID),
			logger.WithValue("event_name", evt.Name))
		return errors.New("circuit breaker is open")
	}

	// Execute with retry
	err := h.executeWithRetry(ctx, evt)

	// Update circuit breaker
	if h.circuitBreaker != nil {
		if err != nil {
			h.circuitBreaker.recordFailure()
		} else {
			h.circuitBreaker.recordSuccess()
		}
	}

	return err
}

// executeWithRetry performs the HTTP request with exponential backoff retry
func (h *HTTPEventDispatcher) executeWithRetry(ctx context.Context, evt *event.Record) error {
	var lastErr error

	for attempt := 0; attempt <= h.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			logger.Of(ctx).WarnS("HTTPEventDispatcher:ContextCancelled",
				logger.WithValue("event_id", evt.ID),
				logger.WithValue("attempt", attempt))
			return ctx.Err()
		default:
		}

		// Calculate backoff delay
		if attempt > 0 {
			delay := h.calculateBackoff(attempt)
			logger.Of(ctx).InfoS("HTTPEventDispatcher:RetryDelay",
				logger.WithValue("event_id", evt.ID),
				logger.WithValue("attempt", attempt),
				logger.WithValue("delay_ms", delay.Milliseconds()))

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Execute request
		err := h.executeRequest(ctx, evt, attempt)
		if err == nil {
			if attempt > 0 {
				logger.Of(ctx).InfoS("HTTPEventDispatcher:RetrySuccess",
					logger.WithValue("event_id", evt.ID),
					logger.WithValue("attempts", attempt+1))
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !h.isRetryableError(err) {
			logger.Of(ctx).ErrorS("HTTPEventDispatcher:NonRetryableError",
				logger.WithValue("event_id", evt.ID),
				logger.WithValue("attempt", attempt),
				logger.WithValue("error", err.Error()))
			return fmt.Errorf("non-retryable error: %w", err)
		}

		logger.Of(ctx).WarnS("HTTPEventDispatcher:RetryableError",
			logger.WithValue("event_id", evt.ID),
			logger.WithValue("attempt", attempt),
			logger.WithValue("error", err.Error()))
	}

	logger.Of(ctx).ErrorS("HTTPEventDispatcher:MaxRetriesExceeded",
		logger.WithValue("event_id", evt.ID),
		logger.WithValue("max_retries", h.config.MaxRetries),
		logger.WithValue("final_error", lastErr.Error()))

	return fmt.Errorf("max retries (%d) exceeded: %w", h.config.MaxRetries, lastErr)
}

// executeRequest performs a single HTTP request
func (h *HTTPEventDispatcher) executeRequest(ctx context.Context, evt *event.Record, attempt int) error {
	// Create request context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, h.config.RequestTimeout)
	defer cancel()

	startTime := time.Now()

	logger.Of(ctx).InfoS("HTTPEventDispatcher:SendingRequest",
		logger.WithValue("event_id", evt.ID),
		logger.WithValue("event_name", evt.Name),
		logger.WithValue("attempt", attempt),
		logger.WithValue("correlation_id", evt.CorrelationID))

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Event-ID":   evt.ID,
	}
	if evt.RequestID != "" {
		headers["X-Request-ID"] = evt.RequestID
	}
	if evt.TraceID != "" {
		headers["X-Trace-ID"] = evt.TraceID
	}
	if evt.SpanID != "" {
		headers["X-Span-ID"] = evt.SpanID
	}

	// Inject standard trace context headers (traceparent/tracestate).
	traceHeaders := http.Header{}
	otel.GetTextMapPropagator().Inject(reqCtx, propagation.HeaderCarrier(traceHeaders))
	for k, values := range traceHeaders {
		if len(values) > 0 && values[0] != "" {
			headers[k] = values[0]
		}
	}

	resp, err := h.client.R().
		SetContext(reqCtx).
		SetBody(evt).
		SetHeaders(headers).
		Post(h.serviceURL)

	duration := time.Since(startTime)

	if err != nil {
		logger.Of(ctx).ErrorS("HTTPEventDispatcher:RequestError",
			logger.WithValue("event_id", evt.ID),
			logger.WithValue("attempt", attempt),
			logger.WithValue("duration_ms", duration.Milliseconds()),
			logger.WithValue("error", err.Error()))
		return fmt.Errorf("request failed: %w", err)
	}

	logger.Of(ctx).InfoS("HTTPEventDispatcher:RequestCompleted",
		logger.WithValue("event_id", evt.ID),
		logger.WithValue("attempt", attempt),
		logger.WithValue("status_code", resp.StatusCode()),
		logger.WithValue("duration_ms", duration.Milliseconds()),
		logger.WithValue("response_size", len(resp.Body())))

	if resp.IsError() {
		errorMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
		logger.Of(ctx).ErrorS("HTTPEventDispatcher:HTTPError",
			logger.WithValue("event_id", evt.ID),
			logger.WithValue("attempt", attempt),
			logger.WithValue("status_code", resp.StatusCode()),
			logger.WithValue("response_body", string(resp.Body())))
		return fmt.Errorf("HTTP error: %s", errorMsg)
	}

	return nil
}

// calculateBackoff calculates the backoff delay with optional jitter
func (h *HTTPEventDispatcher) calculateBackoff(attempt int) time.Duration {
	backoff := float64(h.config.InitialBackoff) * math.Pow(h.config.BackoffFactor, float64(attempt-1))

	if backoff > float64(h.config.MaxBackoff) {
		backoff = float64(h.config.MaxBackoff)
	}

	duration := time.Duration(backoff)

	// Add jitter to prevent thundering herd
	if h.config.JitterEnabled {
		jitter := time.Duration(rand.Float64() * float64(duration) * 0.1) // 10% jitter
		duration += jitter
	}

	return duration
}

// isRetryableError determines if an error should trigger a retry
func (h *HTTPEventDispatcher) isRetryableError(err error) bool {
	// Network errors are retryable
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return netErr.Temporary() || netErr.Timeout()
	}

	// Timeout errors are retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check HTTP status codes (if error contains HTTP info)
	errStr := err.Error()

	// 5xx server errors are retryable
	if contains(errStr, "HTTP 5") {
		return true
	}

	// 429 Too Many Requests is retryable
	if contains(errStr, "HTTP 429") {
		return true
	}

	// 408 Request Timeout is retryable
	if contains(errStr, "HTTP 408") {
		return true
	}

	// 4xx client errors (except specific ones) are not retryable
	if contains(errStr, "HTTP 4") {
		return false
	}

	// Connection errors are retryable
	if contains(errStr, "connection refused") || contains(errStr, "connection reset") {
		return true
	}

	// DNS errors are retryable
	if contains(errStr, "no such host") {
		return true
	}

	// Default to retryable for unknown errors
	return true
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && (func() bool {
			for i := 1; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})()
}

// Circuit breaker methods

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return time.Since(cb.lastFailTime) > cb.config.Timeout
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordSuccess records a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0

	if cb.state == CircuitHalfOpen {
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = CircuitClosed
			cb.successes = 0
		}
	}
}

// recordFailure records a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.state == CircuitClosed && cb.failures >= cb.config.FailureThreshold {
		cb.state = CircuitOpen
	} else if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		cb.successes = 0
	}
}
