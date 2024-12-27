package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/models"
)

func TestDeleteSignatureHandler(t *testing.T) {
	store := models.NewInMemoryDocumentStore()

	// Add a test document
	doc := models.Document{
		SignerName:  "Test User",
		SignerEmail: "test@example.com",
		DeviceID:    "test-device",
		Status:      "pending",
	}
	requestID, err := store.AddDocument(doc)
	if err != nil {
		t.Fatalf("Failed to add test document: %v", err)
	}

	tests := []struct {
		name           string
		requestID      string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "Success",
			requestID:      requestID,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"request_id": requestID,
				"status":     "removed",
			},
		},
		{
			name:           "Not Found",
			requestID:      "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]string{
				"error": "Signature request not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", "/api/documents/signatures/"+tt.requestID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/api/documents/signatures/{request_id}", func(w http.ResponseWriter, r *http.Request) {
				DeleteSignatureHandler(w, r, store)
			}).Methods("DELETE")

			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			var response map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatal("Failed to decode response body")
			}

			for key, expectedValue := range tt.expectedBody {
				if actualValue := response[key]; actualValue != expectedValue {
					t.Errorf("handler returned unexpected body: got %v want %v",
						actualValue, expectedValue)
				}
			}
		})
	}
}
