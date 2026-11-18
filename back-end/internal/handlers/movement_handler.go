package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"odontoestoque/internal/models"
	"odontoestoque/internal/services"
)

type MovementHandler struct {
	service *services.MovementService
	logger  *slog.Logger
}

func NewMovementHandler(service *services.MovementService, logger *slog.Logger) *MovementHandler {
	return &MovementHandler{service: service, logger: logger}
}

// List trata GET /api/movements?tipo=entrada&limite=50.
func (h *MovementHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	movements, err := h.service.List(r.Context(), q.Get("tipo"), q.Get("limite"))
	if err != nil {
		var vErr *models.ValidationError
		if errors.As(err, &vErr) {
			writeError(w, http.StatusUnprocessableEntity, vErr.Error(), vErr.Fields)
			return
		}
		h.logger.Error("erro ao listar movimentações", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno", nil)
		return
	}
	writeJSON(w, http.StatusOK, movements)
}
