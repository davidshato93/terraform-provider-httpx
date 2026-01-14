package utils

import (
	"strings"
)

// RedactHeaders redacts sensitive headers from a map of headers
func RedactHeaders(headers map[string]string, redactList []string) map[string]string {
	if len(redactList) == 0 {
		return headers
	}

	redacted := make(map[string]string)
	redactMap := make(map[string]bool)
	for _, h := range redactList {
		redactMap[strings.ToLower(h)] = true
	}

	for k, v := range headers {
		if redactMap[strings.ToLower(k)] {
			redacted[k] = "[REDACTED]"
		} else {
			redacted[k] = v
		}
	}

	return redacted
}

// RedactHeaderValue redacts a single header value if it's in the redact list
func RedactHeaderValue(headerName string, headerValue string, redactList []string) string {
	if len(redactList) == 0 {
		return headerValue
	}

	for _, h := range redactList {
		if strings.EqualFold(h, headerName) {
			return "[REDACTED]"
		}
	}

	return headerValue
}

// TruncateString truncates a string to maxLen and appends a truncation marker
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... [TRUNCATED]"
}

// RedactError redacts sensitive information from error messages
func RedactError(errMsg string, redactList []string) string {
	// Simple redaction - replace common sensitive patterns
	result := errMsg
	for _, pattern := range redactList {
		// This is a simple implementation; could be enhanced with regex
		if strings.Contains(strings.ToLower(result), strings.ToLower(pattern)) {
			result = strings.ReplaceAll(result, pattern, "[REDACTED]")
		}
	}
	return result
}

