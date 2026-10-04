package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-perpustakaan/app/model"
	"api-perpustakaan/app/repository"
	"api-perpustakaan/helper"
)

// LoanService menangani semua endpoint /loans.
type LoanService struct {
	repo      repository.LoanRepository
	userRepo  repository.UserRepository
	perms     *helper.PermissionSet
}

func NewLoanService(
	repo repository.LoanRepository,
	userRepo repository.UserRepository,
	perms *helper.PermissionSet,
) *LoanService {
	return &LoanService{repo: repo, userRepo: userRepo, perms: perms}
}

func (s *LoanService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	// Jika punya loan:list:any → ambil semua; jika tidak → hanya milik sendiri.
	onlyUserID := current.UserID
	if s.perms.Can(current.Role, "loan:list:any") {
		onlyUserID = 0
	}

	loans, err := s.repo.FindAll(ctx, onlyUserID)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar peminjaman berhasil diambil", loans)
}

func (s *LoanService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	loan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateLoanError(err)
	}

	// Ownership check: peminjam itu sendiri ATAU punya loan:read:any
	if !CanAccessLoan(current, loan.UserID, s.perms) {
		return helper.Forbidden("tidak berhak mengakses peminjaman ini")
	}

	return helper.Success(c, fiber.StatusOK, "peminjaman ditemukan", loan)
}

func (s *LoanService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateLoanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Pastikan peminjam (user_id) ada dan aktif
	borrower, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.BadRequest("peminjam (user_id) tidak ditemukan")
		}
		return helper.Internal(err)
	}
	if !borrower.IsActive {
		return helper.BadRequest("akun peminjam tidak aktif")
	}

	loan := model.Loan{
		BookID:   req.BookID,
		UserID:   req.UserID,
		LoanedBy: current.UserID, // staff yang memproses
		DueDate:  time.Now().AddDate(0, 0, req.DueDays),
	}

	// CreateWithStockDecrement menangani cek stok + insert loan + update stock
	// dalam satu transaksi. ErrOutOfStock diterjemahkan di sini.
	created, err := s.repo.CreateWithStockDecrement(ctx, loan)
	if err != nil {
		return translateLoanError(err)
	}

	// Isi join fields untuk response (tidak ada JOIN dalam transaksi)
	created.Username = borrower.Username

	return helper.Created(c, "peminjaman berhasil dibuat", created,
		"/api/v1/loans/"+strconv.Itoa(created.ID))
}

func (s *LoanService) Return(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	// ReturnWithStockIncrement menangani cek status + update loan + update stock
	// dalam satu transaksi. ErrAlreadyReturned diterjemahkan di sini.
	updated, err := s.repo.ReturnWithStockIncrement(ctx, id)
	if err != nil {
		return translateLoanError(err)
	}

	return helper.Success(c, fiber.StatusOK, "buku berhasil dikembalikan", updated)
}

func (s *LoanService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateLoanError(err)
	}

	return helper.NoContent(c)
}

func translateLoanError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("peminjaman tidak ditemukan")
	case errors.Is(err, repository.ErrBookNotFound):
		return helper.NotFound("buku (book_id) tidak ditemukan")
	case errors.Is(err, repository.ErrOutOfStock):
		return helper.BadRequest("stok buku habis")
	case errors.Is(err, repository.ErrAlreadyReturned):
		return helper.BadRequest("peminjaman sudah dikembalikan")
	default:
		return helper.Internal(err)
	}
}
