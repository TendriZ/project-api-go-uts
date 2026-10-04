package helper

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"api-perpustakaan/app/model"
)

var ErrInvalidCursor = errors.New("cursor tidak valid")

// EncodeBookCursor mengubah pasangan (title, id) menjadi satu string yang aman di URL.
//
// Buku diurutkan title ASC, id ASC. Cursor berisi title dan id (bukan created_at).
// base64 dipakai agar bentuk internalnya dapat kita ubah tanpa mengubah kontrak API.
// Perlu ditegaskan: base64 adalah PENGKODEAN, bukan enkripsi.
func EncodeBookCursor(title string, id int) string {
	raw := url.QueryEscape(title) + "|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeBookCursor membaca kembali penanda dari string.
//
// Seluruh jalur kegagalan mengembalikan error, tidak ada yang "diperbaiki
// diam-diam". Cursor rusak berarti permintaan client memang salah.
func DecodeBookCursor(encoded string) (model.BookCursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.BookCursor{}, ErrInvalidCursor
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.BookCursor{}, ErrInvalidCursor
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return model.BookCursor{}, ErrInvalidCursor
	}

	title, err := url.QueryUnescape(parts[0])
	if err != nil {
		return model.BookCursor{}, ErrInvalidCursor
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return model.BookCursor{}, ErrInvalidCursor
	}

	return model.BookCursor{
		Title: title,
		ID:    id,
	}, nil
}
