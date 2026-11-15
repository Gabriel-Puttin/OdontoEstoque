package handlers

import (
	"log/slog"
	"net/http"

	"odontoestoque/internal/services"
)

type UnitHandler struct {
	service *services.UnitService
	logger  *slog.Logger
}

func NewUnitHandler(service *services.UnitService, logger *slog.Logger) *UnitHandler {
	return &UnitHandler{service: service, logger: logger}
}

// List trata GET /api/units.
func (h *UnitHandler) List(w http.ResponseWriter, r *http.Request) {
	units, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("erro ao listar unidades de medida", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno", nil)
		return
	}
	writeJSON(w, http.StatusOK, units)
}
