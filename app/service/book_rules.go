package service

import (
	"api-perpustakaan/app/model"
)

// ApplyBookPatch menggabungkan perubahan PATCH ke data buku yang ada.
func ApplyBookPatch(current model.Book, req model.PatchBookRequest) model.Book {
	if req.ISBN != nil {
		current.ISBN = *req.ISBN
	}
	if req.Title != nil {
		current.Title = *req.Title
	}
	if req.Author != nil {
		current.Author = *req.Author
	}
	if req.Publisher != nil {
		current.Publisher = *req.Publisher
	}
	if req.Year != nil {
		current.Year = *req.Year
	}
	if req.Stock != nil {
		current.Stock = *req.Stock
	}
	if req.Category != nil {
		current.Category = *req.Category
	}
	return current
}

// IsEmptyBookPatch memeriksa body PATCH yang tidak berisi field apa pun.
func IsEmptyBookPatch(req model.PatchBookRequest) bool {
	return req.ISBN == nil &&
		req.Title == nil &&
		req.Author == nil &&
		req.Publisher == nil &&
		req.Year == nil &&
		req.Stock == nil &&
		req.Category == nil
}
