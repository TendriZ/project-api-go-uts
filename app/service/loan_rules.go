package service

import (
	"api-perpustakaan/app/model"
	"api-perpustakaan/helper"
)

// CanAccessLoan memutuskan apakah seseorang boleh melihat data loan tertentu.
//
// Dua jalur yang diizinkan:
// 1. Peminjam itu sendiri (user_id = current.UserID).
// 2. Punya permission loan:read:any (staff/admin).
func CanAccessLoan(
	current model.AuthUser,
	loanUserID int,
	perms *helper.PermissionSet,
) bool {
	if current.UserID == loanUserID {
		return true
	}
	return perms.Can(current.Role, "loan:read:any")
}
