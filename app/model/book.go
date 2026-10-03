package model

import "time"

// Book merepresentasikan satu baris pada tabel books.
type Book struct {
	ID        int       `json:"id"`
	ISBN      string    `json:"isbn"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Publisher string    `json:"publisher"`
	Year      int       `json:"year"`
	Stock     int       `json:"stock"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateBookRequest adalah body untuk POST /books.
type CreateBookRequest struct {
	ISBN      string `json:"isbn"      validate:"required,min=10,max=17,nospace"`
	Title     string `json:"title"     validate:"required,min=1,max=200"`
	Author    string `json:"author"    validate:"required,min=1,max=150"`
	Publisher string `json:"publisher" validate:"required,min=1,max=150"`
	Year      int    `json:"year"      validate:"required,min=1000,max=2100"`
	Stock     int    `json:"stock"     validate:"min=0"`
	Category  string `json:"category"  validate:"required,min=1,max=50"`
}

// ReplaceBookRequest adalah body untuk PUT /books/:id.
type ReplaceBookRequest struct {
	ISBN      string `json:"isbn"      validate:"required,min=10,max=17,nospace"`
	Title     string `json:"title"     validate:"required,min=1,max=200"`
	Author    string `json:"author"    validate:"required,min=1,max=150"`
	Publisher string `json:"publisher" validate:"required,min=1,max=150"`
	Year      int    `json:"year"      validate:"required,min=1000,max=2100"`
	Stock     int    `json:"stock"     validate:"min=0"`
	Category  string `json:"category"  validate:"required,min=1,max=50"`
}

// PatchBookRequest adalah body untuk PATCH /books/:id.
type PatchBookRequest struct {
	ISBN      *string `json:"isbn,omitempty"      validate:"omitnil,min=10,max=17,nospace"`
	Title     *string `json:"title,omitempty"     validate:"omitnil,min=1,max=200"`
	Author    *string `json:"author,omitempty"    validate:"omitnil,min=1,max=150"`
	Publisher *string `json:"publisher,omitempty" validate:"omitnil,min=1,max=150"`
	Year      *int    `json:"year,omitempty"      validate:"omitnil,min=1000,max=2100"`
	Stock     *int    `json:"stock,omitempty"     validate:"omitnil,min=0"`
	Category  *string `json:"category,omitempty"  validate:"omitnil,min=1,max=50"`
}

// BookCursorQuery adalah parameter untuk cursor pagination buku.
type BookCursorQuery struct {
	Limit    int
	Search   string // pencarian title/author
	Category string // filter kategori
	After    *BookCursor
}

// BookCursor menyimpan posisi cursor untuk buku.
// Buku diurutkan title ASC, id ASC — jadi cursor berisi (title, id).
type BookCursor struct {
	Title string
	ID    int
}
