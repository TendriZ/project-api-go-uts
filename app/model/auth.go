package model

import "time"

// RegisterRequest adalah body untuk POST /auth/register.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
	// TIDAK ADA field Role — mencegah mass assignment ("role":"admin" dari client)
}

// max=72 pada password: bcrypt hanya memproses 72 byte pertama dan mengabaikan
// sisanya tanpa peringatan. Tanpa batas ini, dua password berbeda yang sama
// pada 72 byte pertama akan dianggap identik saat login.

// LoginRequest adalah body untuk POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest adalah body untuk POST /auth/refresh dan POST /auth/logout.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenPair adalah pasangan token yang dikirim ke client setelah login/refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshToken adalah satu baris pada tabel refresh_tokens.
type RefreshToken struct {
	ID        int64
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// AuthUser adalah identitas yang dibawa access token.
// Disimpan di c.Locals setelah RequireAuth berhasil.
type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// CreateUserRequest adalah body untuk POST /users (admin membuat akun).
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
	Role     string `json:"role"     validate:"required,oneof=admin staff user"`
}

// ReplaceUserRequest adalah body untuk PUT /users/:id.
type ReplaceUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	IsActive bool   `json:"is_active"`
}

// PatchUserRequest adalah body untuk PATCH /users/:id.
type PatchUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3,max=30,username"`
	Email    *string `json:"email,omitempty"    validate:"omitnil,email,max=120"`
	Password *string `json:"password,omitempty" validate:"omitnil,max=72,strongpassword"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// AssignRoleRequest adalah body untuk PATCH /users/:id/role.
type AssignRoleRequest struct {
	Role string `json:"role" validate:"required"`
}
