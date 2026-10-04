package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-perpustakaan/app/repository"
	"api-perpustakaan/app/service"
	"api-perpustakaan/config"
	"api-perpustakaan/database"
	"api-perpustakaan/helper"
	"api-perpustakaan/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "api-perpustakaan"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	// Repositories
	userRepository  := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository  := repository.NewRoleRepository(pool)
	bookRepository  := repository.NewBookRepository(pool)
	loanRepository  := repository.NewLoanRepository(pool)

	// Pemetaan role ke permission dibaca SEKALI saat aplikasi menyala.
	// Konsekuensinya: perubahan hak akses di database baru berlaku setelah
	// aplikasi dijalankan ulang. Itu keputusan sadar, bukan kelalaian.
	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	// Services
	authService := service.NewAuthService(
		userRepository, tokenRepository, jwtManager,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
		permissions,
	)
	userService := service.NewUserService(userRepository, permissions)
	bookService := service.NewBookService(bookRepository, permissions)
	loanService := service.NewLoanService(loanRepository, userRepository, permissions)

	app := config.NewApp(logger, route.Dependencies{
		Pool:        pool,
		JWT:         jwtManager,
		Permissions: permissions,
		AuthService: authService,
		UserService: userService,
		BookService: bookService,
		LoanService: loanService,
	})

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("sinyal berhenti diterima, menutup server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	logger.Info("server berhenti dengan rapi")
}
