package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"odontoestoque/internal/models"
	"odontoestoque/internal/services"
)

type ProductHandler struct {
	service *services.ProductService
	logger  *slog.Logger
}

func NewProductHandler(service *services.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{service: service, logger: logger}
}

// List trata GET /api/products.
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.List(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

// Get trata GET /api/products/{id}.
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseProductID(w, r)
	if !ok {
		return
	}
	product, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

// Create trata POST /api/products.
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.ProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	product, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

// Update trata PUT /api/products/{id}.
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseProductID(w, r)
	if !ok {
		return
	}
	var req models.ProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	product, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

// Delete trata DELETE /api/products/{id} (exclusão lógica).
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseProductID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) handleError(w http.ResponseWriter, err error) {
	var vErr *models.ValidationError
	var rErr *models.RuleError
	switch {
	case errors.As(err, &vErr):
		writeError(w, http.StatusUnprocessableEntity, vErr.Error(), vErr.Fields)
	case errors.As(err, &rErr):
		writeError(w, http.StatusUnprocessableEntity, rErr.Message, nil)
	case errors.Is(err, models.ErrProductNotFound):
		writeError(w, http.StatusNotFound, models.ErrProductNotFound.Error(), nil)
	case errors.Is(err, models.ErrProductExists):
		writeError(w, http.StatusConflict, models.ErrProductExists.Error(), nil)
	default:
		h.logger.Error("erro em produtos", "error", err)
		writeError(w, http.StatusInternalServerError, "erro interno", nil)
	}
}

func parseProductID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id inválido", nil)
		return 0, false
	}
	return id, true
}
