package response

import (
	"encoding/json"
	"net/http"
)

type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message"`
	Meta    any    `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type ValidationErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Errors  any    `json:"errors"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func WriteSuccess(w http.ResponseWriter, status int, data any, message string) {
	writeJSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func WriteSuccessWithMeta(w http.ResponseWriter, status int, data any, message string, meta any) {
	writeJSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
		Meta:    meta,
	})
}

func WriteError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func WriteValidationErrors(w http.ResponseWriter, status int, errors any) {
	writeJSON(w, status, ValidationErrorResponse{
		Success: false,
		Message: "validation failed",
		Errors:  errors,
	})
}
