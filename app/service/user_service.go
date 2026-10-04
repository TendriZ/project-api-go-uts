package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-perpustakaan/app/model"
	"api-perpustakaan/app/repository"
	"api-perpustakaan/helper"
)

// UserService menangani semua endpoint /users.
type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	users, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	totalPages := CountTotalPages(total, q.Limit)
	meta := &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return helper.SuccessList(c, "daftar user berhasil diambil", users, meta)
}

func (s *UserService) Get(c *fiber.Ctx) error {
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

	// Cek izin: bisa akses sendiri ATAU punya user:read:any
	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Forbidden("tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

func (s *UserService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	created, err := s.repo.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     req.Role,
		IsActive: true,
	})
	if err != nil {
		return translateUserError(err)
	}

	return helper.Created(c, "user berhasil dibuat", created, "/api/v1/users/"+strconv.Itoa(created.ID))
}

func (s *UserService) Replace(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah data user lain")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	var req model.ReplaceUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, model.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		Password: existing.Password,
		Role:     existing.Role, // role tidak berubah lewat PUT
		IsActive: req.IsActive,
	})
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diganti seluruhnya", result)
}

func (s *UserService) Patch(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah data user lain")
	}

	var req model.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if IsEmptyUserPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	updated := ApplyUserPatch(existing, req)

	// Password diproses terpisah: perlu di-hash
	if req.Password != nil {
		hashed, err := helper.HashPassword(*req.Password)
		if err != nil {
			return helper.Internal(err)
		}
		updated.Password = hashed
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diperbarui sebagian", result)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
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

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	if err := ValidateAssignRole(current, id, req.Role, s.perms); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserError(err)
	}

	existing.Role = req.Role
	result, err := s.repo.Update(ctx, existing)
	if err != nil {
		return translateUserError(err)
	}

	return helper.Success(c, fiber.StatusOK, "role berhasil diubah", result)
}

func (s *UserService) Delete(c *fiber.Ctx) error {
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

	// Tidak bisa hapus akun sendiri
	if current.UserID == id {
		return helper.Forbidden("tidak bisa menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateUserError(err)
	}

	return helper.NoContent(c)
}

func translateUserError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("user tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("username atau email sudah dipakai")
	case errors.Is(err, repository.ErrHasReferencedData):
		return helper.Conflict("tidak dapat menghapus user karena masih memiliki riwayat atau transaksi peminjaman terkait")
	default:
		return helper.Internal(err)
	}
}
