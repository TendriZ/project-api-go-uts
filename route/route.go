package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-perpustakaan/app/service"
	"api-perpustakaan/helper"
	"api-perpustakaan/middleware"
)

// Dependencies menyatukan seluruh dependensi yang dibutuhkan route.
type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	Permissions *helper.PermissionSet
	AuthService *service.AuthService
	UserService *service.UserService
	BookService *service.BookService
	LoanService *service.LoanService
}

// Register mendaftarkan semua route aplikasi.
// File ini HANYA berisi pemetaan URL ke handler — tidak ada logika bisnis di sini.
func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// Publik — tanpa token
	api.Get("/health", healthCheck(deps.Pool))

	perms := deps.Permissions

	// Auth
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// Users — seluruh endpoint wajib token
	// Hak yang dapat diputuskan tanpa data → middleware
	// Hak yang bergantung pada kepemilikan (id sendiri) → service
	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	users.Get("/", middleware.RequirePermission(perms, "user:list"), deps.UserService.List)
	users.Post("/", middleware.RequirePermission(perms, "user:create"), deps.UserService.Create)
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)
	users.Patch("/:id", deps.UserService.Patch)
	users.Delete("/:id", middleware.RequirePermission(perms, "user:delete"), deps.UserService.Delete)

	// Books — GET tidak pakai RequireJSON (juga melayani CSV)
	books := api.Group("/books", middleware.RequireAuth(deps.JWT))
	books.Get("/", middleware.RequirePermission(perms, "book:list"), deps.BookService.List)
	books.Get("/:id", middleware.RequirePermission(perms, "book:read:any"), deps.BookService.Get)
	books.Post("/", middleware.RequireJSON, middleware.RequirePermission(perms, "book:create"), deps.BookService.Create)
	books.Put("/:id", middleware.RequireJSON, middleware.RequirePermission(perms, "book:update:any"), deps.BookService.Replace)
	books.Patch("/:id", middleware.RequireJSON, middleware.RequirePermission(perms, "book:update:any"), deps.BookService.Patch)
	books.Delete("/:id", middleware.RequirePermission(perms, "book:delete"), deps.BookService.Delete)

	// Loans
	loans := api.Group("/loans", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	loans.Get("/", deps.LoanService.List)    // permission dicek di service (loan:list:any vs milik sendiri)
	loans.Post("/", middleware.RequirePermission(perms, "loan:create"), deps.LoanService.Create)
	loans.Get("/:id", deps.LoanService.Get)  // ownership check di service
	loans.Patch("/:id/return", middleware.RequirePermission(perms, "loan:return"), deps.LoanService.Return)
	loans.Delete("/:id", middleware.RequirePermission(perms, "loan:delete"), deps.LoanService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
