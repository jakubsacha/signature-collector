package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/i18n"
	"github.com/jakubsacha/signature-collector/models"
	"github.com/jakubsacha/signature-collector/templates"
)

type DocumentsHandler struct {
	store models.DocumentStore
}

func NewDocumentsHandler(store models.DocumentStore) *DocumentsHandler {
	return &DocumentsHandler{store: store}
}

func (h *DocumentsHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching documents for device: %s", deviceID)
	documents, err := h.store.ListDocuments(deviceID)
	if err != nil {
		log.Printf("Failed to fetch documents: %v", err)
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}
	log.Printf("Found %d pending documents for device %s", len(documents), deviceID)

	// Check if this is a content-only request
	if r.URL.Path == "/documents/"+deviceID+"/content" {
		component := templates.DocumentsContent(deviceID, documents, i18n.T("ConfirmDelete", nil))
		component.Render(r.Context(), w)
		return
	}

	// Full page request
	component := templates.Layout(templates.DocumentsList(deviceID, documents))
	component.Render(r.Context(), w)
}
