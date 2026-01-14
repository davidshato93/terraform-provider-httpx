package provider

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	Attempts            int64
	MinDelayMs          int64
	MaxDelayMs          int64
	Backoff             string
	Jitter              bool
	RetryOnStatusCodes  []int64
	RespectRetryAfter   bool
}

// ShouldRetry determines if a request should be retried based on error or status code
func (rc *RetryConfig) ShouldRetry(err error, statusCode int64) bool {
	// Always retry on transport errors
	if err != nil {
		return true
	}

	// Retry on configured status codes
	for _, code := range rc.RetryOnStatusCodes {
		if statusCode == code {
			return true
		}
	}

	return false
}

// CalculateDelay calculates the delay for the current attempt
func (rc *RetryConfig) CalculateDelay(attempt int64, retryAfter string) time.Duration {
	var delayMs int64

	// Respect Retry-After header if present and enabled
	if rc.RespectRetryAfter && retryAfter != "" {
		if delay, err := parseRetryAfter(retryAfter); err == nil {
			return delay
		}
	}

	// Calculate base delay based on backoff strategy
	switch rc.Backoff {
	case "exponential":
		// Exponential: min_delay * 2^(attempt-1)
		delayMs = rc.MinDelayMs * int64(math.Pow(2, float64(attempt-1)))
	case "linear":
		// Linear: min_delay * attempt
		delayMs = rc.MinDelayMs * attempt
	case "fixed":
		fallthrough
	default:
		// Fixed: always min_delay
		delayMs = rc.MinDelayMs
	}

	// Apply max delay cap
	if delayMs > rc.MaxDelayMs {
		delayMs = rc.MaxDelayMs
	}

	delay := time.Duration(delayMs) * time.Millisecond

	// Apply jitter if enabled (add random 0-25% of delay)
	if rc.Jitter {
		jitterMs := int64(float64(delayMs) * 0.25 * rand.Float64())
		delay += time.Duration(jitterMs) * time.Millisecond
	}

	return delay
}

// parseRetryAfter parses the Retry-After header value
// Supports both seconds (integer) and HTTP-date format
func parseRetryAfter(retryAfter string) (time.Duration, error) {
	retryAfter = strings.TrimSpace(retryAfter)

	// Try parsing as seconds (integer)
	if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	// Try parsing as HTTP-date (RFC 7231)
	// Common formats: "Wed, 21 Oct 2015 07:28:00 GMT", "Wed, 21 Oct 2015 07:28:00 UTC"
	layouts := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.ANSIC,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, retryAfter); err == nil {
			now := time.Now()
			if t.After(now) {
				return t.Sub(now), nil
			}
			return 0, fmt.Errorf("retry-after date is in the past")
		}
	}

	return 0, fmt.Errorf("unable to parse retry-after: %s", retryAfter)
}

// ExecuteRequestWithRetry executes an HTTP request with retry logic
// If retryUntilConfig is provided, it will poll until conditions are met
func ExecuteRequestWithRetry(ctx context.Context, req *http.Request, config *ProviderConfig, retryConfig *RetryConfig, retryUntilConfig *RetryUntilConfig) (*ResponseResult, error) {
	if retryConfig == nil && retryUntilConfig == nil {
		// No retry config, execute once
		return ExecuteRequest(ctx, req, config)
	}

	// If retry_until is configured, we need retry config too
	if retryUntilConfig != nil && retryConfig == nil {
		// Create default retry config for conditional retry
		retryConfig = &RetryConfig{
			Attempts:           60, // Default high for polling
			MinDelayMs:         1000,
			MaxDelayMs:         5000,
			Backoff:            "exponential",
			Jitter:             true,
			RetryOnStatusCodes: []int64{},
			RespectRetryAfter:  true,
		}
	}

	var lastErr error
	var lastResult *ResponseResult
	var retryAfter string

	attempts := retryConfig.Attempts
	if attempts <= 0 {
		attempts = 1 // Default to 1 attempt if not configured
	}

	for attempt := int64(1); attempt <= attempts; attempt++ {
		tflog.Debug(ctx, "Executing HTTP request", map[string]interface{}{
			"attempt": attempt,
			"max_attempts": attempts,
			"url": req.URL.String(),
		})

		// Execute request
		result, err := ExecuteRequest(ctx, req, config)
		if err != nil {
			lastErr = err
			lastResult = result
			
			// Check if we should retry
			if !retryConfig.ShouldRetry(err, 0) || attempt >= attempts {
				return result, err
			}

			// Calculate delay and wait
			delay := retryConfig.CalculateDelay(attempt, "")
			tflog.Debug(ctx, "Request failed, retrying", map[string]interface{}{
				"attempt": attempt,
				"error": err.Error(),
				"delay_ms": delay.Milliseconds(),
			})

			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}

			continue
		}

		// Check conditional retry (retry_until)
		if retryUntilConfig != nil {
			satisfied, unsatisfied := retryUntilConfig.EvaluateRetryUntil(ctx, result)
			if !satisfied && attempt < attempts {
				// Extract Retry-After header if present
				if retryConfig.RespectRetryAfter {
					if retryAfterHeader, ok := result.Headers["Retry-After"]; ok {
						retryAfter = retryAfterHeader
					}
				}

				// Calculate delay and wait
				delay := retryConfig.CalculateDelay(attempt, retryAfter)
				tflog.Debug(ctx, "Conditional retry conditions not met", map[string]interface{}{
					"attempt": attempt,
					"status_code": result.StatusCode,
					"unsatisfied": unsatisfied,
					"delay_ms": delay.Milliseconds(),
				})

				select {
				case <-ctx.Done():
					return result, ctx.Err()
				case <-time.After(delay):
					// Continue to next attempt
				}

				lastResult = result
				continue
			}
			if satisfied {
				// Conditions met, return success
				result.AttemptCount = attempt
				return result, nil
			}
		}

		// Check if we should retry based on status code (only if no retry_until)
		if retryUntilConfig == nil && retryConfig.ShouldRetry(nil, result.StatusCode) && attempt < attempts {
			// Extract Retry-After header if present
			if retryConfig.RespectRetryAfter {
				if retryAfterHeader, ok := result.Headers["Retry-After"]; ok {
					retryAfter = retryAfterHeader
				}
			}

			// Calculate delay and wait
			delay := retryConfig.CalculateDelay(attempt, retryAfter)
			tflog.Debug(ctx, "Status code requires retry", map[string]interface{}{
				"attempt": attempt,
				"status_code": result.StatusCode,
				"delay_ms": delay.Milliseconds(),
			})

			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}

			lastResult = result
			continue
		}

		// Success or no retry needed
		result.AttemptCount = attempt
		return result, nil
	}

	// Exhausted all attempts
	if lastResult != nil {
		lastResult.AttemptCount = attempts
		if retryUntilConfig != nil {
			_, unsatisfied := retryUntilConfig.EvaluateRetryUntil(ctx, lastResult)
			return lastResult, fmt.Errorf("exhausted %d retry attempts, conditions not met: %v", attempts, unsatisfied)
		}
		return lastResult, fmt.Errorf("exhausted %d retry attempts, last status: %d", attempts, lastResult.StatusCode)
	}

	if lastErr != nil {
		return &ResponseResult{
			StatusCode:   0,
			AttemptCount: attempts,
			Error:        lastErr.Error(),
		}, fmt.Errorf("exhausted %d retry attempts: %w", attempts, lastErr)
	}

	return nil, fmt.Errorf("exhausted %d retry attempts", attempts)
}

// BuildRetryConfig converts RetryModel to RetryConfig
func BuildRetryConfig(ctx context.Context, retryModel *RetryModel) *RetryConfig {
	if retryModel == nil {
		return nil
	}

	config := &RetryConfig{
		Attempts:           20, // Default
		MinDelayMs:         250,
		MaxDelayMs:         5000,
		Backoff:            "exponential",
		Jitter:             true,
		RetryOnStatusCodes: []int64{408, 429, 500, 502, 503, 504},
		RespectRetryAfter:  true,
	}

	if !retryModel.Attempts.IsNull() && !retryModel.Attempts.IsUnknown() {
		config.Attempts = retryModel.Attempts.ValueInt64()
	}

	if !retryModel.MinDelayMs.IsNull() && !retryModel.MinDelayMs.IsUnknown() {
		config.MinDelayMs = retryModel.MinDelayMs.ValueInt64()
	}

	if !retryModel.MaxDelayMs.IsNull() && !retryModel.MaxDelayMs.IsUnknown() {
		config.MaxDelayMs = retryModel.MaxDelayMs.ValueInt64()
	}

	if !retryModel.Backoff.IsNull() && !retryModel.Backoff.IsUnknown() {
		backoffStr := retryModel.Backoff.ValueString()
		if backoffStr != "" {
			config.Backoff = backoffStr
		}
	}

	if !retryModel.Jitter.IsNull() && !retryModel.Jitter.IsUnknown() {
		config.Jitter = retryModel.Jitter.ValueBool()
	}

	if !retryModel.RetryOnStatusCodes.IsNull() && !retryModel.RetryOnStatusCodes.IsUnknown() {
		codes, err := ConvertTerraformList(ctx, retryModel.RetryOnStatusCodes, func(v interface{}) (int64, error) {
			if intVal, ok := v.(types.Int64); ok {
				return intVal.ValueInt64(), nil
			}
			return 0, fmt.Errorf("expected int64, got %T", v)
		})
		if err == nil && len(codes) > 0 {
			config.RetryOnStatusCodes = codes
		}
	}

	if !retryModel.RespectRetryAfter.IsNull() && !retryModel.RespectRetryAfter.IsUnknown() {
		config.RespectRetryAfter = retryModel.RespectRetryAfter.ValueBool()
	}

	return config
}

