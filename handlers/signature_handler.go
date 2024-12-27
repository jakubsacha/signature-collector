package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/models"
	"github.com/jakubsacha/signature-collector/templates"
)

type SignatureRequest struct {
	SignatureData string           `json:"signature_data"`
	Consents      []models.Consent `json:"consents"`
}

type SignatureResponse struct {
	Status            string `json:"status"`
	ConsentsProcessed bool   `json:"consents_processed"`
	DeviceID          string `json:"device_id"`
}

type SignatureHandler struct {
	store models.DocumentStore
}

func NewSignatureHandler(store models.DocumentStore) *SignatureHandler {
	return &SignatureHandler{store: store}
}

// ShowSignaturePage handles GET /documents/sign/{request_id}
func (h *SignatureHandler) ShowSignaturePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	// Get document from store
	status, _, err := h.store.GetSignatureStatus(requestID)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// If document is already signed, redirect to the device's documents page
	if status == "completed" {
		http.Error(w, "Document already signed", http.StatusBadRequest)
		return
	}

	// Get document details
	doc, err := h.store.GetDocument(requestID)
	if err != nil {
		log.Printf("Error getting document details: %v", err)
		http.Error(w, "Error getting document details", http.StatusInternalServerError)
		return
	}

	// Render the signature page
	component := templates.Layout(templates.SignaturePage(doc, requestID))
	component.Render(r.Context(), w)
}

// ProcessSignature handles POST /documents/sign/{request_id}
func (h *SignatureHandler) ProcessSignature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	var req SignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get document to verify consents
	doc, err := h.store.GetDocument(requestID)
	if err != nil {
		log.Printf("Error getting document: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Verify all mandatory consents are provided
	for _, section := range doc.DocumentContent {
		if section.Type == "consent" && section.ConsentMandatory != nil && *section.ConsentMandatory {
			consentFound := false
			for _, consent := range req.Consents {
				if consent.ConsentType == *section.ConsentType {
					if !consent.Granted {
						http.Error(w, "Mandatory consent not granted", http.StatusBadRequest)
						return
					}
					consentFound = true
					break
				}
			}
			if !consentFound {
				http.Error(w, "Missing mandatory consent", http.StatusBadRequest)
				return
			}
		}
	}

	// Store signature data and update document status
	if err := h.store.UpdateDocumentSignature(requestID, req.SignatureData); err != nil {
		log.Printf("Error storing signature: %v", err)
		http.Error(w, "Error storing signature", http.StatusInternalServerError)
		return
	}

	// Update document status
	if err := h.store.UpdateDocumentStatus(requestID, "completed"); err != nil {
		log.Printf("Error updating document status: %v", err)
		http.Error(w, "Error updating document status", http.StatusInternalServerError)
		return
	}

	// Store consents
	if err := h.store.StoreConsents(requestID, req.Consents); err != nil {
		log.Printf("Error storing consents: %v", err)
		http.Error(w, "Error storing consents", http.StatusInternalServerError)
		return
	}

	// Send callback if configured
	if doc.CallbackURL != "" {
		go func() {
			log.Printf("Sending callback for document %s", requestID)
			var callbackSender = models.NewCallbackSender()
			if err := callbackSender.SendCallback(doc, req.SignatureData, req.Consents); err != nil {
				// Log the error but don't fail the request
				log.Printf("Error sending callback for document %s: %v", requestID, err)
			}
		}()
	} else {
		log.Printf("No callback URL configured for document %s", requestID)
	}

	response := SignatureResponse{
		Status:            "completed",
		ConsentsProcessed: true,
		DeviceID:          doc.DeviceID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
