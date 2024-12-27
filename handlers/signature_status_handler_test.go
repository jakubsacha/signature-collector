package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/models"
	"github.com/stretchr/testify/assert"
)

func TestSignatureStatusHandler(t *testing.T) {
	store := models.NewInMemoryDocumentStore()

	// Add a document to the store
	docID, _ := store.AddDocument(models.Document{
		DocumentContent: []models.DocumentSection{
			{
				ID:      "section1",
				Type:    "text",
				Content: "https://example.com/doc1.pdf",
			},
		},
		SignerName:  "User One",
		SignerEmail: "user1@example.com",
		DeviceID:    "device_123",
		CallbackURL: "https://client.example.com/callback",
		Status:      "completed",
	})

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/documents/signature-status/{request_id}", func(w http.ResponseWriter, r *http.Request) {
		SignatureStatusHandler(w, r, store)
	})

	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
		expectedResp   *SignatureStatusResponse
	}{
		{
			name:           "Successful signature status retrieval",
			requestID:      docID,
			expectedStatus: http.StatusOK,
			expectedResp: &SignatureStatusResponse{
				RequestID:         docID,
				Status:            "completed",
				SignedDocumentURL: "https://example.com/doc1.pdf",
			},
		},
		{
			name:           "Document not found",
			requestID:      "nonexistent_id",
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/documents/signature-status/"+tt.requestID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response SignatureStatusResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, &response)
			}
		})
	}
}
