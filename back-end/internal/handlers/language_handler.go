package handlers

import (
	"log/slog"
	"net/http"

	"odontoestoque/internal/services"
)

type LanguageHandler struct {
	service *services.LanguageService
	logger  *slog.Logger
}

func NewLanguageHandler(service *services.LanguageService, logger *slog.Logger) *LanguageHandler {
	return &LanguageHandler{service: service, logger: logger}
}

// List trata GET /api/languages.
func (h *LanguageHandler) List(w http.ResponseWriter, r *http.Request) {
	languages, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("erro ao listar idiomas", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno", nil)
		return
	}
	writeJSON(w, http.StatusOK, languages)
}
