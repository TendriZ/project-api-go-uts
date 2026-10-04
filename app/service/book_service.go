package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"api-perpustakaan/app/model"
	"api-perpustakaan/app/repository"
	"api-perpustakaan/helper"
)

// BookService menangani semua endpoint /books.
type BookService struct {
	repo  repository.BookRepository
	perms *helper.PermissionSet
}

func NewBookService(repo repository.BookRepository, perms *helper.PermissionSet) *BookService {
	return &BookService{repo: repo, perms: perms}
}

func (s *BookService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Format dipilih SEBELUM query dijalankan. Bila client meminta format
	// yang tidak dapat kita hasilkan, tidak ada gunanya membebani database.
	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	q, err := helper.ParseBookCursorQuery(c)
	if err != nil {
		return err
	}

	// limit+1: baris tambahan hanya penanda "masih ada halaman berikutnya".
	rows, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	if format == helper.FormatCSV {
		return helper.WriteBooksCSV(c, rows)
	}

	meta := &model.CursorMeta{Limit: q.Limit, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		meta.NextCursor = helper.EncodeBookCursor(last.Title, last.ID)
	}

	return helper.SuccessCursor(c, "daftar buku berhasil diambil", rows, meta)
}

func (s *BookService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateBookError(err)
	}

	return helper.Success(c, fiber.StatusOK, "buku ditemukan", book)
}

func (s *BookService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateBookRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	created, err := s.repo.Create(ctx, model.Book{
		ISBN:      req.ISBN,
		Title:     req.Title,
		Author:    req.Author,
		Publisher: req.Publisher,
		Year:      req.Year,
		Stock:     req.Stock,
		Category:  req.Category,
	})
	if err != nil {
		return translateBookError(err)
	}

	return helper.Created(c, "buku berhasil ditambahkan", created,
		"/api/v1/books/"+strconv.Itoa(created.ID))
}

func (s *BookService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.ReplaceBookRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, model.Book{
		ID:        id,
		ISBN:      req.ISBN,
		Title:     req.Title,
		Author:    req.Author,
		Publisher: req.Publisher,
		Year:      req.Year,
		Stock:     req.Stock,
		Category:  req.Category,
	})
	if err != nil {
		return translateBookError(err)
	}

	return helper.Success(c, fiber.StatusOK, "buku berhasil diganti seluruhnya", result)
}

func (s *BookService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var req model.PatchBookRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if IsEmptyBookPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateBookError(err)
	}

	updated := ApplyBookPatch(current, req)

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateBookError(err)
	}

	return helper.Success(c, fiber.StatusOK, "buku berhasil diperbarui sebagian", result)
}

func (s *BookService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateBookError(err)
	}

	return helper.NoContent(c)
}

func translateBookError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("buku tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("ISBN sudah dipakai")
	case errors.Is(err, repository.ErrHasReferencedData):
		return helper.Conflict("tidak dapat menghapus buku karena masih memiliki riwayat atau transaksi peminjaman terkait")
	default:
		return helper.Internal(err)
	}
}
