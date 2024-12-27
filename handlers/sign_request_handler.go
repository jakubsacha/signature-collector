package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jakubsacha/signature-collector/models"
)

// SignRequest represents the request body for the sign-request endpoint
type SignRequest struct {
	DocumentContent []models.DocumentSection `json:"document_content"`
	DocumentTitle   string                   `json:"document_title"`
	SignerName      string                   `json:"signer_name"`
	SignerEmail     string                   `json:"signer_email"`
	DeviceID        string                   `json:"device_id"`
	CallbackURL     string                   `json:"callback_url"`
}

// SignResponse represents the response body for the sign-request endpoint
type SignResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// SignRequestHandler handles the sign-request endpoint
func SignRequestHandler(w http.ResponseWriter, r *http.Request, store models.DocumentStore) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SignerName == "" || req.SignerEmail == "" || req.DeviceID == "" || req.CallbackURL == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Add the document to the database
	doc := models.Document{
		DocumentContent: req.DocumentContent,
		DocumentTitle:   req.DocumentTitle,
		SignerName:      req.SignerName,
		SignerEmail:     req.SignerEmail,
		DeviceID:        req.DeviceID,
		CallbackURL:     req.CallbackURL,
		Status:          "pending",
	}

	requestID, err := store.AddDocument(doc)
	if err != nil {
		log.Printf("Error adding document: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := SignResponse{
		RequestID: requestID,
		Status:    "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
