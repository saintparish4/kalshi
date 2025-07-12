package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	APIResponse
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(message string) *APIResponse {
	return &APIResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now(),
	}
}

// NewMessageResponse creates a message API response
func NewMessageResponse(message string) *APIResponse {
	return &APIResponse{
		Success:   true,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// WithRequestID adds request ID to response
func (r *APIResponse) WithRequestID(requestID string) *APIResponse {
	r.RequestID = requestID
	return r
}

// WriteResponse writes API response to HTTP response writer
func WriteResponse(w http.ResponseWriter, status int, response *APIResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponse writes successful response
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) error {
	return WriteResponse(w, http.StatusOK, NewSuccessResponse(data))
}

// WriteErrorResponse writes error response
func WriteErrorResponse(w http.ResponseWriter, status int, message string) error {
	return WriteResponse(w, status, NewErrorResponse(message))
}

// WriteCreatedResponse writes created response
func WriteCreatedResponse(w http.ResponseWriter, data interface{}) error {
	return WriteResponse(w, http.StatusCreated, NewSuccessResponse(data))
}

// WriteNoContentResponse writes no content response
func WriteNoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WritePaginatedResponse writes paginated response
func WritePaginatedResponse(w http.ResponseWriter, data interface{}, pagination PaginationInfo) error {
	response := &PaginatedResponse{
		APIResponse: APIResponse{
			Success:   true,
			Data:      data,
			Timestamp: time.Now(),
		},
		Pagination: pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

// CalculatePagination calculates pagination info
func CalculatePagination(page, perPage, total int) PaginationInfo {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// StatusCodeToMessage converts HTTP status code to message
func StatusCodeToMessage(code int) string {
	switch code {
	case http.StatusOK:
		return "OK"
	case http.StatusCreated:
		return "Created"
	case http.StatusAccepted:
		return "Accepted"
	case http.StatusNoContent:
		return "No Content"
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not Found"
	case http.StatusMethodNotAllowed:
		return "Method Not Allowed"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusTooManyRequests:
		return "Too Many Requests"
	case http.StatusInternalServerError:
		return "Internal Server Error"
	case http.StatusBadGateway:
		return "Bad Gateway"
	case http.StatusServiceUnavailable:
		return "Service Unavailable"
	default:
		return "Unknown Status"
	}
}
