package service

import (
	"api-perpustakaan/app/model"
)

// ApplyUserPatch menggabungkan perubahan PATCH ke data user yang ada.
// Pemeriksaan bentuk sudah selesai dikerjakan tag sebelum fungsi ini dipanggil.
func ApplyUserPatch(current model.User, req model.PatchUserRequest) model.User {
	if req.Username != nil {
		current.Username = *req.Username
	}
	if req.Email != nil {
		current.Email = *req.Email
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	// Password diproses terpisah di service (perlu hash)
	return current
}

// IsEmptyUserPatch memeriksa body PATCH yang tidak berisi field apa pun.
func IsEmptyUserPatch(req model.PatchUserRequest) bool {
	return req.Username == nil &&
		req.Email == nil &&
		req.Password == nil &&
		req.IsActive == nil
}

// CountTotalPages menghitung jumlah halaman total untuk offset pagination.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return pages
}
