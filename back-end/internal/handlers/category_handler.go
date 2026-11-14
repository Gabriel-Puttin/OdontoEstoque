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

type CategoryHandler struct {
	service *services.CategoryService
	logger  *slog.Logger
}

func NewCategoryHandler(service *services.CategoryService, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{service: service, logger: logger}
}

// List trata GET /api/categories.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

// Products trata GET /api/categories/{id}/products.
func (h *CategoryHandler) Products(w http.ResponseWriter, r *http.Request) {
	id, ok := parseCategoryID(w, r)
	if !ok {
		return
	}
	resp, err := h.service.Products(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Create trata POST /api/categories.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CategoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	category, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, category)
}

// Update trata PUT /api/categories/{id}.
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseCategoryID(w, r)
	if !ok {
		return
	}
	var req models.CategoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	category, err := h.service.Rename(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, category)
}

// Delete trata DELETE /api/categories/{id}.
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseCategoryID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) handleError(w http.ResponseWriter, err error) {
	var vErr *models.ValidationError
	switch {
	case errors.As(err, &vErr):
		writeError(w, http.StatusUnprocessableEntity, vErr.Error(), vErr.Fields)
	case errors.Is(err, models.ErrCategoryNotFound):
		writeError(w, http.StatusNotFound, models.ErrCategoryNotFound.Error(), nil)
	case errors.Is(err, models.ErrCategoryNameExists):
		writeError(w, http.StatusConflict, models.ErrCategoryNameExists.Error(), nil)
	case errors.Is(err, models.ErrCategoryHasProducts):
		writeError(w, http.StatusConflict, models.ErrCategoryHasProducts.Error(), nil)
	default:
		h.logger.Error("erro em categorias", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno", nil)
	}
}

// parseCategoryID lê o {id} da URL (categorias.id é smallint).
func parseCategoryID(w http.ResponseWriter, r *http.Request) (int16, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 16)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id inválido", nil)
		return 0, false
	}
	return int16(id), true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido", nil)
		return false
	}
	return true
}
