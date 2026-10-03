package model

import "time"

// LoanStatus merepresentasikan status peminjaman buku.
type LoanStatus string

const (
	StatusBorrowed LoanStatus = "borrowed"
	StatusReturned LoanStatus = "returned"
)

// Loan merepresentasikan satu baris pada tabel loans.
type Loan struct {
	ID         int        `json:"id"`
	BookID     int        `json:"book_id"`
	BookTitle  string     `json:"book_title"` // JOIN dari books
	UserID     int        `json:"user_id"`     // peminjam
	Username   string     `json:"username"`    // JOIN dari users
	LoanedBy   int        `json:"loaned_by"`   // staff yang memproses
	LoanDate   time.Time  `json:"loan_date"`
	DueDate    time.Time  `json:"due_date"`
	ReturnDate *time.Time `json:"return_date,omitempty"`
	Status     LoanStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateLoanRequest adalah body untuk POST /loans.
// staff/admin membuat peminjaman atas nama peminjam (user_id).
type CreateLoanRequest struct {
	BookID  int `json:"book_id"  validate:"required,min=1"`
	UserID  int `json:"user_id"  validate:"required,min=1"` // id peminjam
	DueDays int `json:"due_days" validate:"required,min=1,max=30"`
}
