package service

import (
	"api-perpustakaan/app/model"
	"api-perpustakaan/helper"
)

// CanAccessUser memutuskan apakah seseorang boleh menyentuh data user tertentu.
//
// Dua jalur yang diizinkan:
// 1. Kepemilikan — id pemanggil sama dengan target.
// 2. Permission — role-nya memang berhak atas data siapa pun (:any).
//
// Fungsi ini murni: tidak mengimpor fiber maupun repository.
func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

// ValidateAssignRole memeriksa apakah operasi role:assign sah.
//
// Dua kondisi yang ditolak:
// 1. Mengubah role diri sendiri.
// 2. Role yang diminta tidak dikenal sistem.
func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	newRole string,
	perms *helper.PermissionSet,
) error {
	if current.UserID == targetID {
		return helper.Forbidden("tidak bisa mengubah role diri sendiri")
	}
	if !perms.IsKnownRole(newRole) {
		return helper.BadRequest("role tidak valid: " + newRole)
	}
	return nil
}
