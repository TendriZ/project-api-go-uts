package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-perpustakaan/app/model"
)

// Sentinel errors dipakai seluruh repository dalam package ini.
// Didefinisikan SEKALI di sini — tidak boleh diduplikasi di file lain.
var (
	ErrNotFound          = errors.New("data tidak ditemukan")
	ErrDuplicate         = errors.New("data sudah ada")
	ErrHasReferencedData = errors.New("data memiliki keterkaitan relasi data lain")
)

// isUniqueViolation memeriksa apakah error adalah unique constraint violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// isForeignKeyViolation memeriksa apakah error adalah foreign key constraint violation (23503).
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

// UserRepository mendefinisikan kontrak akses data untuk user.
type UserRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.User, int, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	Create(ctx context.Context, u model.User) (model.User, error)
	Update(ctx context.Context, u model.User) (model.User, error)
	Delete(ctx context.Context, id int) error
}

var allowedUserSortCols = map[string]string{
	"id":         "id",
	"username":   "username",
	"email":      "email",
	"created_at": "created_at",
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.User, int, error) {
	where := " WHERE 1=1"
	args := []any{}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", len(args), len(args))
	}

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung user: %w", err)
	}

	col, ok := allowedUserSortCols[q.Sort]
	if !ok {
		col = "id"
	}
	arah := "ASC"
	if q.Order == "desc" {
		arah = "DESC"
	}

	sqlText := fmt.Sprintf(
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users%s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		where, col, arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	hasil := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris user: %w", err)
		}
		hasil = append(hasil, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query user: %w", err)
	}
	return hasil, total, nil
}

func (r *userPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

// FindByUsername dipakai saat login. Tidak membedakan huruf besar/kecil,
// sama seperti unique index-nya.
func (r *userPostgresRepository) FindByUsername(
	ctx context.Context, username string,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role,
		&u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Create(
	ctx context.Context, u model.User,
) (model.User, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		u.Username, u.Email, u.Password, u.Role, u.IsActive,
	).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Update(
	ctx context.Context, u model.User,
) (model.User, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET username=$1, email=$2, is_active=$3, role=$4
		 WHERE id=$5
		 RETURNING id, username, email, password, role, is_active, created_at`,
		u.Username, u.Email, u.IsActive, u.Role, u.ID,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("memperbarui user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrHasReferencedData
		}
		return fmt.Errorf("menghapus user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
