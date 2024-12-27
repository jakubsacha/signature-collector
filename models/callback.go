package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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

// makeCallbackRequest attempts a single callback request
func (s *CallbackSender) makeCallbackRequest(url string, jsonData []byte) error {
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending callback: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Callback request successful with status: %d", resp.StatusCode)
		return nil
	}

	return fmt.Errorf("callback request failed with status: %d", resp.StatusCode)
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
		return fmt.Errorf("error marshaling callback payload: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt < s.cfg.maxRetries; attempt++ {
		log.Printf("Sending callback for document %s, attempt %d", doc.ID, attempt)
		err := s.makeCallbackRequest(doc.CallbackURL, jsonData)
		if err == nil {
			return nil
		}

		var backoff = calculateBackoff(attempt, s.cfg)
		log.Printf("Callback request failed with error: %v, waiting %s before next attempt", err, backoff)
		lastErr = err

		// Wait before next retry
		s.sleepFunc(backoff)
	}

	return fmt.Errorf("failed after %d retries. Last error: %v", s.cfg.maxRetries, lastErr)
}
