package models

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallbackSender_SendCallback(t *testing.T) {
	// Mock time for consistent testing
	mockTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTimeFunc := func() time.Time {
		return mockTime
	}

	tests := []struct {
		name           string
		doc            Document
		signatureData  string
		consents       []Consent
		serverBehavior func(w http.ResponseWriter)
		expectedError  string
		retryConfig    *retryConfig
		sleepCalls     int
	}{
		{
			name: "successful callback",
			doc: Document{
				ID:          "123",
				SignerName:  "John Doe",
				SignerEmail: "john@example.com",
				Status:      "completed",
				CallbackURL: "http://example.com/callback",
			},
			signatureData: "signature123",
			consents: []Consent{
				{ConsentType: "marketing", Granted: true},
			},
			serverBehavior: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusOK)
			},
			sleepCalls: 0,
		},
		{
			name: "missing callback URL",
			doc: Document{
				ID:          "123",
				SignerName:  "John Doe",
				SignerEmail: "john@example.com",
				Status:      "completed",
			},
			expectedError: "no callback URL provided",
			sleepCalls:    0,
		},
		{
			name: "server error with retry",
			doc: Document{
				ID:          "123",
				SignerName:  "John Doe",
				SignerEmail: "john@example.com",
				Status:      "completed",
				CallbackURL: "http://example.com/callback",
			},
			serverBehavior: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			retryConfig: &retryConfig{
				maxRetries: 3,
				baseDelay:  10 * time.Millisecond,
				maxDelay:   100 * time.Millisecond,
			},
			expectedError: "failed after 3 retries",
			sleepCalls:    3,
		},
		{
			name: "eventual success after retries",
			doc: Document{
				ID:          "123",
				SignerName:  "John Doe",
				SignerEmail: "john@example.com",
				Status:      "completed",
				CallbackURL: "http://example.com/callback",
			},
			serverBehavior: (func() func(w http.ResponseWriter) {
				attempts := 0
				return func(w http.ResponseWriter) {
					attempts++
					if attempts < 2 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			})(),
			retryConfig: &retryConfig{
				maxRetries: 3,
				baseDelay:  10 * time.Millisecond,
				maxDelay:   100 * time.Millisecond,
			},
			sleepCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server if callback URL is provided
			var ts *httptest.Server
			if tt.doc.CallbackURL != "" {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					tt.serverBehavior(w)
				}))
				defer ts.Close()
				tt.doc.CallbackURL = ts.URL
			}

			// Track sleep calls
			sleepCalls := 0
			mockSleep := func(d time.Duration) {
				sleepCalls++
			}

			// Create sender with mocked dependencies
			sender := NewCallbackSender()
			sender.timeNow = mockTimeFunc
			sender.sleepFunc = mockSleep

			if tt.retryConfig != nil {
				sender.WithRetryConfig(*tt.retryConfig)
			}

			// Execute test
			err := sender.SendCallback(tt.doc, tt.signatureData, tt.consents)

			// Verify results
			if tt.expectedError != "" {
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.sleepCalls, sleepCalls, "unexpected number of sleep calls")
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	cfg := retryConfig{
		baseDelay: 100 * time.Millisecond,
		maxDelay:  1 * time.Second,
	}

	tests := []struct {
		name          string
		attempt       int
		expectedDelay time.Duration
	}{
		{
			name:          "first attempt",
			attempt:       0,
			expectedDelay: 100 * time.Millisecond,
		},
		{
			name:          "second attempt",
			attempt:       1,
			expectedDelay: 200 * time.Millisecond,
		},
		{
			name:          "third attempt",
			attempt:       2,
			expectedDelay: 400 * time.Millisecond,
		},
		{
			name:          "max delay reached",
			attempt:       5,
			expectedDelay: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateBackoff(tt.attempt, cfg)
			assert.Equal(t, tt.expectedDelay, delay)
		})
	}
}

func TestCallbackSender_WithClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	sender := NewCallbackSender().WithClient(customClient)
	assert.Equal(t, customClient, sender.client)
}

func TestCallbackSender_WithRetryConfig(t *testing.T) {
	customConfig := retryConfig{
		maxRetries: 5,
		baseDelay:  50 * time.Millisecond,
		maxDelay:   500 * time.Millisecond,
	}

	sender := NewCallbackSender().WithRetryConfig(customConfig)
	assert.Equal(t, customConfig, sender.cfg)
}
