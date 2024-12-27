package handlers

import (
	"net/http"

	"github.com/jakubsacha/signature-collector/templates"
)

type DeviceEntryHandler struct{}

func NewDeviceEntryHandler() *DeviceEntryHandler {
	return &DeviceEntryHandler{}
}

func (h *DeviceEntryHandler) ShowForm(w http.ResponseWriter, r *http.Request) {
	component := templates.Layout(templates.DeviceIDForm())
	component.Render(r.Context(), w)
}

func (h *DeviceEntryHandler) ProcessForm(w http.ResponseWriter, r *http.Request) {
	deviceID := r.FormValue("device_id")
	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	// Set HX-Redirect header for HTMX to handle the redirect
	w.Header().Set("HX-Redirect", "/documents/"+deviceID)
	w.WriteHeader(http.StatusOK)
}
