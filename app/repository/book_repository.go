package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-perpustakaan/app/model"
)

// BookRepository mendefinisikan kontrak akses data untuk buku.
type BookRepository interface {
	FindAll(ctx context.Context, q model.BookCursorQuery) ([]model.Book, error)
	FindByID(ctx context.Context, id int) (model.Book, error)
	Create(ctx context.Context, b model.Book) (model.Book, error)
	Update(ctx context.Context, b model.Book) (model.Book, error)
	Delete(ctx context.Context, id int) error
}

type bookPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) BookRepository {
	return &bookPostgresRepository{pool: pool}
}

// FindAll mengambil satu halaman buku memakai keyset pagination.
//
// Buku diurutkan title ASC, id ASC. Cursor condition adalah (title, id) > ($n, $m)
// karena urutan ASC — berbeda dari pola student yang memakai < (DESC).
// id sebagai tiebreaker menjamin tidak ada baris yang terlewat atau terkirim dua kali.
// Jumlah yang diminta sengaja limit+1; baris tambahan hanya penanda "has_more".
func (r *bookPostgresRepository) FindAll(
	ctx context.Context,
	q model.BookCursorQuery,
) ([]model.Book, error) {
	args := []any{}
	where := " WHERE 1=1"

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (title ILIKE $%d OR author ILIKE $%d)", len(args), len(args))
	}

	if q.Category != "" {
		args = append(args, q.Category)
		where += fmt.Sprintf(" AND category = $%d", len(args))
	}

	// PENTING: cursor condition menggunakan > karena urutan ASC (title ASC, id ASC).
	// Jangan gunakan < seperti pada student (created_at DESC, id DESC).
	if q.After != nil {
		args = append(args, q.After.Title, q.After.ID)
		where += fmt.Sprintf(
			" AND (title, id) > ($%d, $%d)",
			len(args)-1,
			len(args),
		)
	}

	args = append(args, q.Limit+1)

	query := fmt.Sprintf(
		`SELECT id, isbn, title, author, publisher, year, stock, category, created_at
		 FROM books%s
		 ORDER BY title ASC, id ASC
		 LIMIT $%d`,
		where,
		len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar buku: %w", err)
	}
	defer rows.Close()

	result := []model.Book{}
	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.Publisher,
			&b.Year, &b.Stock, &b.Category, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("membaca baris buku: %w", err)
		}
		result = append(result, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query buku: %w", err)
	}

	return result, nil
}

func (r *bookPostgresRepository) FindByID(ctx context.Context, id int) (model.Book, error) {
	var b model.Book
	err := r.pool.QueryRow(ctx,
		`SELECT id, isbn, title, author, publisher, year, stock, category, created_at
		 FROM books WHERE id = $1`, id,
	).Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.Publisher,
		&b.Year, &b.Stock, &b.Category, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Book{}, ErrNotFound
		}
		return model.Book{}, fmt.Errorf("mengambil buku: %w", err)
	}
	return b, nil
}

func (r *bookPostgresRepository) Create(ctx context.Context, b model.Book) (model.Book, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO books (isbn, title, author, publisher, year, stock, category)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at`,
		b.ISBN, b.Title, b.Author, b.Publisher, b.Year, b.Stock, b.Category,
	).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Book{}, ErrDuplicate
		}
		return model.Book{}, fmt.Errorf("menyimpan buku: %w", err)
	}
	return b, nil
}

func (r *bookPostgresRepository) Update(ctx context.Context, b model.Book) (model.Book, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE books SET isbn=$1, title=$2, author=$3, publisher=$4, year=$5, stock=$6, category=$7
		 WHERE id=$8
		 RETURNING id, isbn, title, author, publisher, year, stock, category, created_at`,
		b.ISBN, b.Title, b.Author, b.Publisher, b.Year, b.Stock, b.Category, b.ID,
	).Scan(&b.ID, &b.ISBN, &b.Title, &b.Author, &b.Publisher,
		&b.Year, &b.Stock, &b.Category, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Book{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Book{}, ErrDuplicate
		}
		return model.Book{}, fmt.Errorf("memperbarui buku: %w", err)
	}
	return b, nil
}

func (r *bookPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrHasReferencedData
		}
		return fmt.Errorf("menghapus buku: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
