package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/models"
)

// DeleteSignatureResponse represents the response body for the delete-signature endpoint
type DeleteSignatureResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// DeleteSignatureHandler handles the delete-signature endpoint
func DeleteSignatureHandler(w http.ResponseWriter, r *http.Request, store models.DocumentStore) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	// Get the document to verify it exists
	_, err := store.GetDocument(requestID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Signature request not found",
		})
		return
	}

	// Update document status to removed
	err = store.UpdateDocumentStatus(requestID, "removed")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := DeleteSignatureResponse{
		RequestID: requestID,
		Status:    "removed",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
