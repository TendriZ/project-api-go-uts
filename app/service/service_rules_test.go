package service_test

import (
	"testing"

	"api-perpustakaan/app/model"
	"api-perpustakaan/app/service"
	"api-perpustakaan/helper"
)

func TestApplyUserPatch(t *testing.T) {
	current := model.User{
		ID:       1,
		Username: "lama",
		Email:    "lama@perpus.id",
		IsActive: true,
	}

	newUsername := "baru"
	newActive := false
	patched := service.ApplyUserPatch(current, model.PatchUserRequest{
		Username: &newUsername,
		IsActive: &newActive,
	})

	if patched.Username != "baru" {
		t.Errorf("expected username 'baru', got '%s'", patched.Username)
	}
	if patched.Email != "lama@perpus.id" {
		t.Errorf("expected email unchanged, got '%s'", patched.Email)
	}
	if patched.IsActive != false {
		t.Errorf("expected isActive false, got %v", patched.IsActive)
	}
}

func TestIsEmptyUserPatch(t *testing.T) {
	empty := model.PatchUserRequest{}
	if !service.IsEmptyUserPatch(empty) {
		t.Error("expected IsEmptyUserPatch to return true for empty patch")
	}

	val := "halo"
	nonEmpty := model.PatchUserRequest{Username: &val}
	if service.IsEmptyUserPatch(nonEmpty) {
		t.Error("expected IsEmptyUserPatch to return false for non-empty patch")
	}
}

func TestCountTotalPages(t *testing.T) {
	tests := []struct {
		total int
		limit int
		want  int
	}{
		{total: 0, limit: 10, want: 0},
		{total: 10, limit: 10, want: 1},
		{total: 11, limit: 10, want: 2},
		{total: 25, limit: 10, want: 3},
		{total: 50, limit: 0, want: 0},
	}

	for _, tt := range tests {
		got := service.CountTotalPages(tt.total, tt.limit)
		if got != tt.want {
			t.Errorf("CountTotalPages(%d, %d) = %d; want %d", tt.total, tt.limit, got, tt.want)
		}
	}
}

func TestApplyBookPatch(t *testing.T) {
	current := model.Book{
		ID:        1,
		ISBN:      "978-602-03-8591-5",
		Title:     "Bumi Manusia",
		Author:    "Pramoedya Ananta Toer",
		Publisher: "Hasta Mitra",
		Year:      1980,
		Stock:     5,
		Category:  "Sastra",
	}

	newTitle := "Bumi Manusia Edisi Khusus"
	newStock := 10
	patched := service.ApplyBookPatch(current, model.PatchBookRequest{
		Title: &newTitle,
		Stock: &newStock,
	})

	if patched.Title != newTitle {
		t.Errorf("expected title '%s', got '%s'", newTitle, patched.Title)
	}
	if patched.Stock != 10 {
		t.Errorf("expected stock 10, got %d", patched.Stock)
	}
	if patched.Author != "Pramoedya Ananta Toer" {
		t.Errorf("expected author unchanged, got '%s'", patched.Author)
	}
}

func TestIsEmptyBookPatch(t *testing.T) {
	empty := model.PatchBookRequest{}
	if !service.IsEmptyBookPatch(empty) {
		t.Error("expected IsEmptyBookPatch to return true for empty struct")
	}

	cat := "Teknologi"
	nonEmpty := model.PatchBookRequest{Category: &cat}
	if service.IsEmptyBookPatch(nonEmpty) {
		t.Error("expected IsEmptyBookPatch to return false for non-empty struct")
	}
}

func TestCanAccessLoan(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"staff": {"loan:read:any"},
		"user":  {"book:list"},
	})

	borrower := model.AuthUser{UserID: 42, Role: "user"}
	otherUser := model.AuthUser{UserID: 99, Role: "user"}
	staffUser := model.AuthUser{UserID: 10, Role: "staff"}

	// 1. Borrower mengakses miliknya sendiri -> boleh
	if !service.CanAccessLoan(borrower, 42, perms) {
		t.Error("borrower should be able to access own loan")
	}

	// 2. User lain mengakses loan milik orang lain -> dilarang
	if service.CanAccessLoan(otherUser, 42, perms) {
		t.Error("other user should NOT be able to access loan")
	}

	// 3. Staff mengakses loan milik siapa pun (memiliki loan:read:any) -> boleh
	if !service.CanAccessLoan(staffUser, 42, perms) {
		t.Error("staff with loan:read:any should be able to access any loan")
	}
}

func TestCanAccessUser(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"user:read:any", "user:update:any"},
		"user":  {},
	})

	owner := model.AuthUser{UserID: 5, Role: "user"}
	stranger := model.AuthUser{UserID: 8, Role: "user"}
	admin := model.AuthUser{UserID: 1, Role: "admin"}

	// Pemilik sendiri
	if !service.CanAccessUser(owner, 5, perms, "user:read:any") {
		t.Error("owner should be able to access own data")
	}

	// Orang lain tanpa permission
	if service.CanAccessUser(stranger, 5, perms, "user:read:any") {
		t.Error("stranger without permission should NOT access data")
	}

	// Admin dengan permission :any
	if !service.CanAccessUser(admin, 5, perms, "user:read:any") {
		t.Error("admin with user:read:any should access data")
	}
}

func TestValidateAssignRole(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {},
		"staff": {},
		"user":  {},
	})

	caller := model.AuthUser{UserID: 1, Role: "admin"}

	// Skenario 1: Mengubah role diri sendiri -> DILARANG
	errSelf := service.ValidateAssignRole(caller, 1, "staff", perms)
	if errSelf == nil {
		t.Error("expected error when admin changes own role")
	}

	// Skenario 2: Mengubah ke role yang tidak valid -> DILARANG
	errUnknown := service.ValidateAssignRole(caller, 2, "supergod", perms)
	if errUnknown == nil {
		t.Error("expected error for unknown role")
	}

	// Skenario 3: Mengubah role user lain ke role valid -> SUKSES
	errValid := service.ValidateAssignRole(caller, 2, "staff", perms)
	if errValid != nil {
		t.Errorf("expected nil error for valid assign, got %v", errValid)
	}
}
