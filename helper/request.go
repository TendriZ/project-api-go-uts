package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-perpustakaan/app/model"
)

// RequestContext memberi timeout untuk setiap operasi database.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID membaca parameter :id dari jalur dan memastikan bentuknya benar.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

var allowedUserSort = map[string]bool{
	"id":         true,
	"username":   true,
	"email":      true,
	"created_at": true,
}

// ParseListQuery membaca query string dan memberi nilai bawaan yang aman (untuk users).
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}

	if q.Limit < 1 {
		q.Limit = 10
	}

	if q.Limit > 100 {
		q.Limit = 100
	}

	if !allowedUserSort[q.Sort] {
		q.Sort = "id"
	}

	if q.Order != "desc" {
		q.Order = "asc"
	}

	return q
}

// ParseBookCursorQuery membaca query string untuk pagination buku berbasis cursor.
// Cursor rusak adalah kesalahan pemakai API (400), bukan nilai yang diam-diam diperbaiki.
func ParseBookCursorQuery(c *fiber.Ctx) (model.BookCursorQuery, error) {
	q := model.BookCursorQuery{
		Limit:    c.QueryInt("limit", 10),
		Search:   strings.TrimSpace(c.Query("search")),
		Category: strings.TrimSpace(c.Query("category")),
	}

	if q.Limit < 1 {
		q.Limit = 10
	}

	if q.Limit > 100 {
		q.Limit = 100
	}

	if raw := strings.TrimSpace(c.Query("cursor")); raw != "" {
		cursor, err := DecodeBookCursor(raw)
		if err != nil {
			return model.BookCursorQuery{}, BadRequest("cursor tidak valid")
		}

		q.After = &cursor
	}

	return q, nil
}
