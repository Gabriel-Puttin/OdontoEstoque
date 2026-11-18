package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido", nil)
		return
	}

	user, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			writeError(w, http.StatusForbidden, err.Error(), nil)
			return
		}
		h.logger.Error("falha no login", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno", nil)
		return
	}

	writeJSON(w, http.StatusOK, models.LoginResponse{
		ID:           user.ID,
		Nome:         user.Nome,
		Email:        user.Email,
		IdiomaCodigo: user.IdiomaCodigo,
	})
}

// Update trata PUT /api/users/{id}.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id inválido", nil)
		return
	}

	var req models.UpdateUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		var vErr *models.ValidationError
		switch {
		case errors.As(err, &vErr):
			writeError(w, http.StatusUnprocessableEntity, vErr.Error(), vErr.Fields)
		case errors.Is(err, models.ErrUserNotFound):
			writeError(w, http.StatusNotFound, models.ErrUserNotFound.Error(), nil)
		case errors.Is(err, models.ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, models.ErrEmailAlreadyExists.Error(), nil)
		case errors.Is(err, models.ErrInvalidLanguage):
			writeError(w, http.StatusUnprocessableEntity, models.ErrInvalidLanguage.Error(), map[string]string{"idioma_codigo": "não existe"})
		default:
			h.logger.Error("falha ao atualizar usuário", "error", err)
			writeError(w, http.StatusInternalServerError, "erro interno", nil)
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "logout realizado"})
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
