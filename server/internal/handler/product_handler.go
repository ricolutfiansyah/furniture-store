package handler

import (
	"errors"
	"furniture-api/internal/response"
	"furniture-api/internal/service"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) GetAllProduct(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if limit > 100 {
		limit = 100
	}

	result, err := h.productService.GetAll(r.Context(), page, limit)
	if err != nil {
		log.Printf("get all products: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccessWithMeta(w, http.StatusOK, result.Products, "products retrieved successfully", response.PaginationMeta{
		Page:  result.Page,
		Limit: result.PageSize,
		Total: result.Total,
	})
}

func (h *ProductHandler) GetProductBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.WriteError(w, http.StatusBadRequest, "product slug is required")
		return
	}

	product, err := h.productService.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response.WriteError(w, http.StatusNotFound, "product not found")
			return
		}
		log.Printf("get product by slug %q: %v", slug, err)
		response.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, product, "product retrieved successfully")
}
