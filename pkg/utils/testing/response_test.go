package testing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kalshi/pkg/utils"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"message": "success"}
	response := utils.NewSuccessResponse(data)

	if !response.Success {
		t.Error("NewSuccessResponse() should set Success to true")
	}

	if response.Data == nil {
		t.Error("NewSuccessResponse() should set Data field")
	}

	if response.Error != "" {
		t.Error("NewSuccessResponse() should not set Error field")
	}

	if response.Timestamp.IsZero() {
		t.Error("NewSuccessResponse() should set Timestamp")
	}
}

func TestNewErrorResponse(t *testing.T) {
	message := "Something went wrong"
	response := utils.NewErrorResponse(message)

	if response.Success {
		t.Error("NewErrorResponse() should set Success to false")
	}

	if response.Error != message {
		t.Errorf("NewErrorResponse() Error = %v, want %v", response.Error, message)
	}

	if response.Data != nil {
		t.Error("NewErrorResponse() should not set Data field")
	}

	if response.Timestamp.IsZero() {
		t.Error("NewErrorResponse() should set Timestamp")
	}
}

func TestNewMessageResponse(t *testing.T) {
	message := "Operation completed"
	response := utils.NewMessageResponse(message)

	if !response.Success {
		t.Error("NewMessageResponse() should set Success to true")
	}

	if response.Message != message {
		t.Errorf("NewMessageResponse() Message = %v, want %v", response.Message, message)
	}

	if response.Data != nil {
		t.Error("NewMessageResponse() should not set Data field")
	}

	if response.Timestamp.IsZero() {
		t.Error("NewMessageResponse() should set Timestamp")
	}
}

func TestAPIResponse_WithRequestID(t *testing.T) {
	response := utils.NewSuccessResponse("test")
	requestID := "req-123"

	result := response.WithRequestID(requestID)

	if result.RequestID != requestID {
		t.Errorf("WithRequestID() RequestID = %v, want %v", result.RequestID, requestID)
	}

	// Should return the same response instance
	if result != response {
		t.Error("WithRequestID() should return the same response instance")
	}
}

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	response := utils.NewSuccessResponse("test data")

	err := utils.WriteResponse(w, http.StatusOK, response)
	if err != nil {
		t.Errorf("WriteResponse() unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("WriteResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteResponse() Content-Type = %v, want %v", contentType, "application/json")
	}
}

func TestWriteSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := "test data"

	err := utils.WriteSuccessResponse(w, data)
	if err != nil {
		t.Errorf("WriteSuccessResponse() unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("WriteSuccessResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteSuccessResponse() Content-Type = %v, want %v", contentType, "application/json")
	}
}

func TestWriteErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	message := "Bad request"

	err := utils.WriteErrorResponse(w, http.StatusBadRequest, message)
	if err != nil {
		t.Errorf("WriteErrorResponse() unexpected error: %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("WriteErrorResponse() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteErrorResponse() Content-Type = %v, want %v", contentType, "application/json")
	}
}

func TestWriteCreatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := "created data"

	err := utils.WriteCreatedResponse(w, data)
	if err != nil {
		t.Errorf("WriteCreatedResponse() unexpected error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("WriteCreatedResponse() status = %v, want %v", w.Code, http.StatusCreated)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WriteCreatedResponse() Content-Type = %v, want %v", contentType, "application/json")
	}
}

func TestWriteNoContentResponse(t *testing.T) {
	w := httptest.NewRecorder()

	utils.WriteNoContentResponse(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("WriteNoContentResponse() status = %v, want %v", w.Code, http.StatusNoContent)
	}
}

func TestWritePaginatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"item1", "item2", "item3"}
	pagination := utils.PaginationInfo{
		Page:       1,
		PerPage:    10,
		Total:      25,
		TotalPages: 3,
		HasNext:    true,
		HasPrev:    false,
	}

	err := utils.WritePaginatedResponse(w, data, pagination)
	if err != nil {
		t.Errorf("WritePaginatedResponse() unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("WritePaginatedResponse() status = %v, want %v", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("WritePaginatedResponse() Content-Type = %v, want %v", contentType, "application/json")
	}
}

func TestCalculatePagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		perPage  int
		total    int
		expected utils.PaginationInfo
	}{
		{
			name:    "normal pagination",
			page:    2,
			perPage: 10,
			total:   25,
			expected: utils.PaginationInfo{
				Page:       2,
				PerPage:    10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    true,
			},
		},
		{
			name:    "first page",
			page:    1,
			perPage: 10,
			total:   25,
			expected: utils.PaginationInfo{
				Page:       1,
				PerPage:    10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    false,
			},
		},
		{
			name:    "last page",
			page:    3,
			perPage: 10,
			total:   25,
			expected: utils.PaginationInfo{
				Page:       3,
				PerPage:    10,
				Total:      25,
				TotalPages: 3,
				HasNext:    false,
				HasPrev:    true,
			},
		},
		{
			name:    "invalid page",
			page:    0,
			perPage: 10,
			total:   25,
			expected: utils.PaginationInfo{
				Page:       1,
				PerPage:    10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    false,
			},
		},
		{
			name:    "invalid per page",
			page:    1,
			perPage: 0,
			total:   25,
			expected: utils.PaginationInfo{
				Page:       1,
				PerPage:    10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    false,
			},
		},
		{
			name:    "empty result",
			page:    1,
			perPage: 10,
			total:   0,
			expected: utils.PaginationInfo{
				Page:       1,
				PerPage:    10,
				Total:      0,
				TotalPages: 1,
				HasNext:    false,
				HasPrev:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CalculatePagination(tt.page, tt.perPage, tt.total)

			if result.Page != tt.expected.Page {
				t.Errorf("CalculatePagination() Page = %v, want %v", result.Page, tt.expected.Page)
			}
			if result.PerPage != tt.expected.PerPage {
				t.Errorf("CalculatePagination() PerPage = %v, want %v", result.PerPage, tt.expected.PerPage)
			}
			if result.Total != tt.expected.Total {
				t.Errorf("CalculatePagination() Total = %v, want %v", result.Total, tt.expected.Total)
			}
			if result.TotalPages != tt.expected.TotalPages {
				t.Errorf("CalculatePagination() TotalPages = %v, want %v", result.TotalPages, tt.expected.TotalPages)
			}
			if result.HasNext != tt.expected.HasNext {
				t.Errorf("CalculatePagination() HasNext = %v, want %v", result.HasNext, tt.expected.HasNext)
			}
			if result.HasPrev != tt.expected.HasPrev {
				t.Errorf("CalculatePagination() HasPrev = %v, want %v", result.HasPrev, tt.expected.HasPrev)
			}
		})
	}
}

func TestStatusCodeToMessage(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{http.StatusOK, "OK"},
		{http.StatusCreated, "Created"},
		{http.StatusAccepted, "Accepted"},
		{http.StatusNoContent, "No Content"},
		{http.StatusBadRequest, "Bad Request"},
		{http.StatusUnauthorized, "Unauthorized"},
		{http.StatusForbidden, "Forbidden"},
		{http.StatusNotFound, "Not Found"},
		{http.StatusMethodNotAllowed, "Method Not Allowed"},
		{http.StatusConflict, "Conflict"},
		{http.StatusTooManyRequests, "Too Many Requests"},
		{http.StatusInternalServerError, "Internal Server Error"},
		{http.StatusBadGateway, "Bad Gateway"},
		{http.StatusServiceUnavailable, "Service Unavailable"},
		{999, "Unknown Status"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := utils.StatusCodeToMessage(tt.code)
			if result != tt.expected {
				t.Errorf("StatusCodeToMessage(%d) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}
