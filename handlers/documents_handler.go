package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakubsacha/signature-collector/i18n"
	"github.com/jakubsacha/signature-collector/logging"
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

	logger := logging.WithField("device_id", deviceID)
	logger.Info("Fetching documents for device")
	documents, err := h.store.ListDocuments(deviceID)
	if err != nil {
		logger.WithField("error", err.Error()).Error("Failed to fetch documents")
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}
	logger.WithField("count", len(documents)).Info("Found pending documents for device")

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
