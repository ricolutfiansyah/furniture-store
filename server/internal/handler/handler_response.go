package handler

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeSuccess(w http.ResponseWriter, status int, data any, message string) {
	writeJSON(w, status, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}
