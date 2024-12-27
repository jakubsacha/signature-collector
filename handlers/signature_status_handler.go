package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/models"
)

// SignatureStatusResponse represents the response body for the signature-status endpoint
type SignatureStatusResponse struct {
	RequestID         string `json:"request_id"`
	Status            string `json:"status"`
	SignedDocumentURL string `json:"signed_document_url,omitempty"`
}

// SignatureStatusHandler handles the signature-status endpoint
func SignatureStatusHandler(w http.ResponseWriter, r *http.Request, store models.DocumentStore) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	// Get the signature status from the store
	status, signedDocumentURL, err := store.GetSignatureStatus(requestID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := SignatureStatusResponse{
		RequestID:         requestID,
		Status:            status,
		SignedDocumentURL: signedDocumentURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
