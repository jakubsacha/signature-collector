package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jakubsacha/signature-collector/logging"
)

// CallbackPayload represents the data sent to the callback URL
type CallbackPayload struct {
	RequestID     string    `json:"request_id"`
	Status        string    `json:"status"`
	SignerName    string    `json:"signer_name"`
	SignerEmail   string    `json:"signer_email"`
	SignatureData string    `json:"signature_data"`
	Consents      []Consent `json:"consents"`
	CompletedAt   time.Time `json:"completed_at"`
}

// retryConfig holds configuration for retry behavior
type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// CallbackSender handles sending callbacks with configurable behavior
type CallbackSender struct {
	client    *http.Client
	cfg       retryConfig
	timeNow   func() time.Time
	sleepFunc func(time.Duration)
}

// NewCallbackSender creates a new CallbackSender with default configuration
func NewCallbackSender() *CallbackSender {
	return &CallbackSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cfg: retryConfig{
			maxRetries: 60,
			baseDelay:  100 * time.Millisecond,
			maxDelay:   30 * time.Second,
		},
		timeNow:   time.Now,
		sleepFunc: time.Sleep,
	}
}

// WithClient sets a custom HTTP client
func (s *CallbackSender) WithClient(client *http.Client) *CallbackSender {
	s.client = client
	return s
}

// WithRetryConfig sets custom retry configuration
func (s *CallbackSender) WithRetryConfig(cfg retryConfig) *CallbackSender {
	s.cfg = cfg
	return s
}

// extractHost extracts the host from a URL string for logging purposes
func extractHost(callbackURL string) string {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return "unknown (parse error)"
	}
	return parsedURL.Host
}

// makeCallbackRequest attempts a single callback request
func (s *CallbackSender) makeCallbackRequest(callbackURL string, jsonData []byte) error {
	host := extractHost(callbackURL)
	resp, err := s.client.Post(callbackURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logging.WithFields(map[string]interface{}{
			"host":        host,
			"status_code": resp.StatusCode,
		}).Info("Callback request successful")
		return nil
	}

	return fmt.Errorf("HTTP %d", resp.StatusCode)
}

// calculateBackoff determines the delay for the next retry attempt
func calculateBackoff(attempt int, cfg retryConfig) time.Duration {
	delay := time.Duration(1<<uint(attempt)) * cfg.baseDelay
	if delay > cfg.maxDelay {
		return cfg.maxDelay
	}
	return delay
}

// SendCallback sends a POST request to the callback URL with signature details
func (s *CallbackSender) SendCallback(doc Document, signatureData string, consents []Consent) error {
	if doc.CallbackURL == "" {
		return fmt.Errorf("no callback URL provided")
	}

	host := extractHost(doc.CallbackURL)
	logger := logging.WithFields(map[string]interface{}{
		"document_id":  doc.ID,
		"host":         host,
		"callback_url": doc.CallbackURL,
	})
	logger.Info("Initiating callback")

	payload := CallbackPayload{
		RequestID:     doc.ID,
		Status:        doc.Status,
		SignerName:    doc.SignerName,
		SignerEmail:   doc.SignerEmail,
		SignatureData: signatureData,
		Consents:      consents,
		CompletedAt:   s.timeNow(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to marshal callback payload")
		return fmt.Errorf("error marshaling callback payload: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt < s.cfg.maxRetries; attempt++ {
		attemptLogger := logger.WithFields(map[string]interface{}{
			"attempt":     attempt + 1,
			"max_retries": s.cfg.maxRetries,
		})
		attemptLogger.Info("Sending callback attempt")

		err := s.makeCallbackRequest(doc.CallbackURL, jsonData)
		if err == nil {
			attemptLogger.Info("Callback completed successfully")
			return nil
		}

		backoff := calculateBackoff(attempt, s.cfg)
		attemptLogger.WithFields(map[string]interface{}{
			"error":         err.Error(),
			"next_retry_in": backoff.String(),
		}).Warn("Callback attempt failed")
		lastErr = err

		// Wait before next retry
		s.sleepFunc(backoff)
	}

	logger.WithFields(map[string]interface{}{
		"total_attempts": s.cfg.maxRetries,
		"last_error":     lastErr.Error(),
	}).Error("Callback failed after all retries")

	return fmt.Errorf("callback to host %s failed after %d retries: %v", host, s.cfg.maxRetries, lastErr)
}
