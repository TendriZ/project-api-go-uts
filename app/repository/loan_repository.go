package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-perpustakaan/app/model"
)

// LoanRepository mendefinisikan kontrak akses data untuk peminjaman buku.
type LoanRepository interface {
	// FindAll mengambil daftar loan. onlyUserID=0 berarti ambil semua (staff/admin).
	FindAll(ctx context.Context, onlyUserID int) ([]model.Loan, error)
	FindByID(ctx context.Context, id int) (model.Loan, error)
	// CreateWithStockDecrement menggunakan transaksi: insert loan + kurangi stock buku.
	CreateWithStockDecrement(ctx context.Context, loan model.Loan) (model.Loan, error)
	// ReturnWithStockIncrement menggunakan transaksi: update loan + tambah stock buku.
	ReturnWithStockIncrement(ctx context.Context, id int) (model.Loan, error)
	Delete(ctx context.Context, id int) error
}

const loanJoinColumns = `
	l.id, l.book_id, b.title, l.user_id, u.username,
	l.loaned_by, l.loan_date, l.due_date, l.return_date, l.status, l.created_at`

type loanPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewLoanRepository(pool *pgxpool.Pool) LoanRepository {
	return &loanPostgresRepository{pool: pool}
}

func (r *loanPostgresRepository) FindAll(
	ctx context.Context,
	onlyUserID int,
) ([]model.Loan, error) {
	where := " WHERE 1=1"
	args := []any{}

	if onlyUserID > 0 {
		args = append(args, onlyUserID)
		where += fmt.Sprintf(" AND l.user_id = $%d", len(args))
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM loans l
		JOIN books b ON b.id = l.book_id
		JOIN users u ON u.id = l.user_id
		%s
		ORDER BY l.created_at DESC, l.id DESC`,
		loanJoinColumns, where,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar peminjaman: %w", err)
	}
	defer rows.Close()

	result := []model.Loan{}
	for rows.Next() {
		l, err := scanLoan(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris peminjaman: %w", err)
		}
		result = append(result, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query peminjaman: %w", err)
	}

	return result, nil
}

func (r *loanPostgresRepository) FindByID(ctx context.Context, id int) (model.Loan, error) {
	var l model.Loan
	err := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s
		 FROM loans l
		 JOIN books b ON b.id = l.book_id
		 JOIN users u ON u.id = l.user_id
		 WHERE l.id = $1`, loanJoinColumns),
		id,
	).Scan(&l.ID, &l.BookID, &l.BookTitle, &l.UserID, &l.Username,
		&l.LoanedBy, &l.LoanDate, &l.DueDate, &l.ReturnDate, &l.Status, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Loan{}, ErrNotFound
		}
		return model.Loan{}, fmt.Errorf("mengambil peminjaman: %w", err)
	}
	return l, nil
}

// CreateWithStockDecrement membuat loan dan mengurangi stok dalam satu transaksi pgx.
//
// Urutan operasi dalam transaksi:
//  1. Ambil stok buku (SELECT ... FOR UPDATE — mengunci baris agar tidak ada race condition)
//  2. Tolak jika stok habis
//  3. INSERT loan
//  4. UPDATE books SET stock = stock - 1
//  5. COMMIT
//
// Bila langkah mana pun gagal, transaksi di-rollback secara otomatis oleh defer.
func (r *loanPostgresRepository) CreateWithStockDecrement(
	ctx context.Context,
	loan model.Loan,
) (model.Loan, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Loan{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Kunci baris buku dan baca stok — FOR UPDATE mencegah race condition
	//    jika dua request pinjam buku yang sama bersamaan.
	var stock int
	err = tx.QueryRow(ctx,
		`SELECT stock FROM books WHERE id = $1 FOR UPDATE`,
		loan.BookID,
	).Scan(&stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Loan{}, ErrBookNotFound
		}
		return model.Loan{}, fmt.Errorf("mengambil stok buku: %w", err)
	}

	// 2. Tolak jika stok habis
	if stock <= 0 {
		return model.Loan{}, ErrOutOfStock
	}

	// 3. Insert loan
	err = tx.QueryRow(ctx,
		`INSERT INTO loans (book_id, user_id, loaned_by, due_date)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, loan_date, created_at`,
		loan.BookID, loan.UserID, loan.LoanedBy, loan.DueDate,
	).Scan(&loan.ID, &loan.LoanDate, &loan.CreatedAt)
	if err != nil {
		return model.Loan{}, fmt.Errorf("menyimpan peminjaman: %w", err)
	}
	loan.Status = model.StatusBorrowed

	// 4. Kurangi stok — operasi ini satu paket dengan insert loan
	_, err = tx.Exec(ctx,
		`UPDATE books SET stock = stock - 1 WHERE id = $1`,
		loan.BookID,
	)
	if err != nil {
		return model.Loan{}, fmt.Errorf("mengurangi stok buku: %w", err)
	}

	// 5. Commit — jika gagal, defer Rollback akan membersihkan
	if err := tx.Commit(ctx); err != nil {
		return model.Loan{}, fmt.Errorf("commit transaksi peminjaman: %w", err)
	}

	return loan, nil
}

// ReturnWithStockIncrement menandai loan sebagai returned dan menambah stok dalam satu transaksi pgx.
//
// Urutan operasi:
//  1. Ambil dan kunci baris loan (SELECT ... FOR UPDATE)
//  2. Tolak jika sudah returned
//  3. UPDATE loan (return_date = NOW(), status = 'returned')
//  4. UPDATE books SET stock = stock + 1
//  5. COMMIT
func (r *loanPostgresRepository) ReturnWithStockIncrement(
	ctx context.Context,
	id int,
) (model.Loan, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Loan{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Kunci dan baca loan
	var loan model.Loan
	err = tx.QueryRow(ctx,
		`SELECT id, book_id, user_id, loaned_by, loan_date, due_date, return_date, status, created_at
		 FROM loans WHERE id = $1 FOR UPDATE`,
		id,
	).Scan(&loan.ID, &loan.BookID, &loan.UserID, &loan.LoanedBy,
		&loan.LoanDate, &loan.DueDate, &loan.ReturnDate, &loan.Status, &loan.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Loan{}, ErrNotFound
		}
		return model.Loan{}, fmt.Errorf("mengambil peminjaman: %w", err)
	}

	// 2. Tolak jika sudah dikembalikan
	if loan.Status == model.StatusReturned {
		return model.Loan{}, ErrAlreadyReturned
	}

	// 3. Update loan
	now := time.Now()
	err = tx.QueryRow(ctx,
		`UPDATE loans SET return_date = $1, status = 'returned' WHERE id = $2
		 RETURNING return_date, status`,
		now, id,
	).Scan(&loan.ReturnDate, &loan.Status)
	if err != nil {
		return model.Loan{}, fmt.Errorf("memperbarui peminjaman: %w", err)
	}

	// 4. Tambah stok buku — satu paket dengan update loan
	_, err = tx.Exec(ctx,
		`UPDATE books SET stock = stock + 1 WHERE id = $1`,
		loan.BookID,
	)
	if err != nil {
		return model.Loan{}, fmt.Errorf("menambah stok buku: %w", err)
	}

	// 5. Commit
	if err := tx.Commit(ctx); err != nil {
		return model.Loan{}, fmt.Errorf("commit transaksi pengembalian: %w", err)
	}

	return loan, nil
}

func (r *loanPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM loans WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus peminjaman: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// scanLoan membaca satu baris dari pgx.Rows ke model.Loan.
func scanLoan(row pgx.Row) (model.Loan, error) {
	var l model.Loan
	err := row.Scan(
		&l.ID, &l.BookID, &l.BookTitle, &l.UserID, &l.Username,
		&l.LoanedBy, &l.LoanDate, &l.DueDate, &l.ReturnDate, &l.Status, &l.CreatedAt,
	)
	return l, err
}
