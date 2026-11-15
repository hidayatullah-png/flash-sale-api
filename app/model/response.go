package model

// ErrorResponse adalah format error standar API.
type ErrorResponse struct {
	Error string `json:"error" example:"produk tidak ditemukan"`
}

// MessageResponse adalah format pesan sukses standar API.
type MessageResponse struct {
	Message string `json:"message" example:"berhasil dihapus"`
}
