package repository

import "errors"

// Sentinel errors khusus operasi loan.
// ErrNotFound dan ErrDuplicate sudah didefinisikan di user_repository.go.
var (
	// ErrOutOfStock dilempar saat stok buku habis saat peminjaman.
	ErrOutOfStock = errors.New("stok buku habis")
	// ErrAlreadyReturned dilempar saat buku sudah dikembalikan sebelumnya.
	ErrAlreadyReturned = errors.New("peminjaman sudah dikembalikan")
	// ErrBookNotFound dilempar saat buku yang dipinjam tidak ada di database.
	ErrBookNotFound = errors.New("buku tidak ditemukan")
)

