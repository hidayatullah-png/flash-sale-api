package model

// ErrorResponse adalah format error standar API (Bawaan Cia).
type ErrorResponse struct {
	Error string `json:"error" example:"produk tidak ditemukan"`
}

// MessageResponse adalah format pesan sukses standar API (Bawaan Cia).
type MessageResponse struct {
	Message string `json:"message" example:"berhasil dihapus"`
}

// ====================================================================
// FORMAT RESPONSE STANDAR GLOBAL (Tambahkan ini agar Helper tidak error)
// ====================================================================

type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

type APIErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}