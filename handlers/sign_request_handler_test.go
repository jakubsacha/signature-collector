package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jakubsacha/signature-collector/models"
	"github.com/stretchr/testify/assert"
)

func TestSignRequestHandler(t *testing.T) {
	store := models.NewInMemoryDocumentStore()

	consentGranted := true
	consentMandatory := true
	consentDefault := false

	validDocumentContent := []models.DocumentSection{
		{
			ID:      "section1",
			Type:    "text",
			Content: "This is the document content.",
		},
		{
			ID:               "section2",
			Type:             "consent",
			Content:          "Marketing consent",
			ConsentType:      stringPtr("marketing_email"),
			ConsentGranted:   &consentGranted,
			ConsentMandatory: &consentMandatory,
			ConsentDefault:   &consentDefault,
		},
	}

	tests := []struct {
		name           string
		method         string
		body           interface{}
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:   "Valid Request",
			method: http.MethodPost,
			body: SignRequest{
				DocumentContent: validDocumentContent,
				SignerName:      "Test User",
				SignerEmail:     "test@example.com",
				DeviceID:        "test_device_id",
				CallbackURL:     "https://client.example.com/callback",
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Invalid Method",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid Body",
			method:         http.MethodPost,
			body:           "invalid body",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing Required Fields",
			method: http.MethodPost,
			body: SignRequest{
				DocumentContent: validDocumentContent,
				SignerName:      "", // Missing required field
				SignerEmail:     "test@example.com",
				DeviceID:        "test_device_id",
				CallbackURL:     "https://client.example.com/callback",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/api/documents/sign-request", bytes.NewReader(body))
			w := httptest.NewRecorder()

			SignRequestHandler(w, req, store)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse {
				var response SignResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.RequestID)
				assert.Equal(t, "pending", response.Status)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
