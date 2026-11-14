package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"odontoestoque/internal/models"
	"odontoestoque/internal/services"
)

type UserHandler struct {
	service *services.UserService
	logger  *slog.Logger
}

func NewUserHandler(service *services.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

// Create trata POST /api/v1/users.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req models.CreateUserRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido", nil)
		return
	}

	user, err := h.service.Create(r.Context(), req)
	if err != nil {
		var vErr *models.ValidationError
		switch {
		case errors.As(err, &vErr):
			writeError(w, http.StatusUnprocessableEntity, vErr.Error(), vErr.Fields)
		case errors.Is(err, models.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, err.Error(), nil)
		case errors.Is(err, models.ErrInvalidLanguage):
			writeError(w, http.StatusUnprocessableEntity, err.Error(), map[string]string{"idioma_codigo": "não existe"})
		default:
			h.logger.Error("falha ao criar usuário", "error", err)
			writeError(w, http.StatusInternalServerError, "erro interno", nil)
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// ---------- helpers de resposta ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, fields map[string]string) {
	body := map[string]any{"error": msg}
	if len(fields) > 0 {
		body["fields"] = fields
	}
	writeJSON(w, status, body)
}
