package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"robot-hand-server/internal/auth"
	"robot-hand-server/internal/config"
	"robot-hand-server/internal/httpapi"
	"robot-hand-server/internal/httpapi/handlers"
	"robot-hand-server/internal/httpapi/middleware"
	"robot-hand-server/internal/mailer"
	"robot-hand-server/internal/repository"
	"robot-hand-server/internal/service"
	"robot-hand-server/internal/storage"
)

type passwordResetMailer interface {
	SendPasswordResetCode(ctx context.Context, email, code string) error
}

func Run() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRepo, revokedTokenRepo, passwordResetCodeRepo, ticketRepo, cleanup, err := initRepositories(ctx, cfg)
	if err != nil {
		log.Fatalf("repository init failed: %v", err)
	}
	defer cleanup()

	resetMailer, supportMailer, err := initMailers(cfg)
	if err != nil {
		log.Fatalf("mailer init failed: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.TokenTTL)
	imageStore := storage.NewLocalImageStore(cfg.UploadDir)

authService := service.NewAuthService(
		userRepo,
		revokedTokenRepo,
		passwordResetCodeRepo,
		jwtManager,
		resetMailer,
		cfg.PasswordResetCodeTTL,
		cfg.JWTSecret,
	)
	supportService := service.NewSupportService(userRepo, ticketRepo, imageStore, supportMailer)

	// === СНАЧАЛА СОЗДАЕМ SYNC REPO ===
	syncRepo := repository.NewInMemorySyncRepository()
	syncService := service.NewSyncService(syncRepo)

	// === ТЕПЕРЬ ПЕРЕДАЕМ ЕГО В PROFILE SERVICE ===
	profileService := service.NewProfileService(userRepo, syncRepo)

	authHandler := handlers.NewAuthHandler(authService)
	supportHandler := handlers.NewSupportHandler(supportService)
	profileHandler := handlers.NewProfileHandler(profileService)
	
	// === НАШ НОВЫЙ ХЕНДЛЕР ===
	syncHandler := handlers.NewSyncHandler(syncService)
	
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// === ДОБАВИЛИ syncHandler В РОУТЕР ===
	router := httpapi.NewRouter(authHandler, supportHandler, profileHandler, syncHandler, authMiddleware)

	srv := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server started on %s", cfg.HTTPAddress)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func initRepositories(
	ctx context.Context,
	cfg config.Config,
) (
	repository.UserRepository,
	repository.RevokedTokenRepository,
	repository.PasswordResetCodeRepository,
	repository.SupportTicketRepository,
	func(),
	error,
) {
	driver := strings.ToLower(strings.TrimSpace(cfg.StorageDriver))

	switch driver {
	case "", "auto":
		if strings.TrimSpace(cfg.DatabaseURL) == "" {
			log.Printf("DATABASE_URL is empty, using in-memory repositories")
			return newInMemoryRepositories()
		}

		userRepo, revokedRepo, resetRepo, ticketRepo, cleanup, err := newPostgresRepositories(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Printf("postgres unavailable, using in-memory repositories: %v", err)
			return newInMemoryRepositories()
		}

		log.Printf("using postgres repositories")
		return userRepo, revokedRepo, resetRepo, ticketRepo, cleanup, nil
	case "memory":
		log.Printf("using in-memory repositories")
		return newInMemoryRepositories()
	case "postgres":
		if strings.TrimSpace(cfg.DatabaseURL) == "" {
			return nil, nil, nil, nil, nil, fmt.Errorf("DATABASE_URL is required when STORAGE_DRIVER=postgres")
		}

		userRepo, revokedRepo, resetRepo, ticketRepo, cleanup, err := newPostgresRepositories(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		log.Printf("using postgres repositories")
		return userRepo, revokedRepo, resetRepo, ticketRepo, cleanup, nil
	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported STORAGE_DRIVER=%q", cfg.StorageDriver)
	}
}

func newPostgresRepositories(
	ctx context.Context,
	databaseURL string,
) (
	repository.UserRepository,
	repository.RevokedTokenRepository,
	repository.PasswordResetCodeRepository,
	repository.SupportTicketRepository,
	func(),
	error,
) {
	db, err := repository.OpenPostgres(ctx, databaseURL)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	if err := repository.EnsureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, nil, nil, nil, nil, err
	}

	cleanup := func() {
		_ = db.Close()
	}

	return repository.NewPostgresUserRepository(db),
		repository.NewPostgresRevokedTokenRepository(db),
		repository.NewPostgresPasswordResetCodeRepository(db),
		repository.NewPostgresSupportTicketRepository(db),
		cleanup,
		nil
}

func newInMemoryRepositories() (
	repository.UserRepository,
	repository.RevokedTokenRepository,
	repository.PasswordResetCodeRepository,
	repository.SupportTicketRepository,
	func(),
	error,
) {
	return repository.NewInMemoryUserRepository(),
		repository.NewInMemoryRevokedTokenRepository(),
		repository.NewInMemoryPasswordResetCodeRepository(),
		repository.NewInMemorySupportTicketRepository(),
		func() {},
		nil
}

func initMailers(cfg config.Config) (passwordResetMailer, mailer.SupportMailer, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.MailerDriver))

	switch driver {
	case "", "auto":
		if hasSMTPConfig(cfg) {
			smtpMailer, err := mailer.NewSMTPMailer(
				cfg.SMTPHost,
				cfg.SMTPPort,
				cfg.SMTPUsername,
				cfg.SMTPPassword,
				cfg.SMTPFrom,
				cfg.SupportRecipientEmail,
			)
			if err == nil {
				log.Printf("using smtp mailer")
				return smtpMailer, smtpMailer, nil
			}
			log.Printf("smtp init failed, using log mailers: %v", err)
		} else {
			log.Printf("smtp config not set, using log mailers")
		}
		return mailer.NewLogPasswordResetMailer(), mailer.NewLogSupportMailer(), nil
	case "log":
		log.Printf("using log mailers")
		return mailer.NewLogPasswordResetMailer(), mailer.NewLogSupportMailer(), nil
	case "noop":
		log.Printf("mail delivery disabled")
		return nil, nil, nil
	case "smtp":
		smtpMailer, err := mailer.NewSMTPMailer(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFrom,
			cfg.SupportRecipientEmail,
		)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("using smtp mailer")
		return smtpMailer, smtpMailer, nil
	default:
		return nil, nil, fmt.Errorf("unsupported MAILER_DRIVER=%q", cfg.MailerDriver)
	}
}

func hasSMTPConfig(cfg config.Config) bool {
	return strings.TrimSpace(cfg.SMTPHost) != "" &&
		cfg.SMTPPort > 0 &&
		strings.TrimSpace(cfg.SMTPUsername) != "" &&
		strings.TrimSpace(cfg.SMTPPassword) != "" &&
		strings.TrimSpace(cfg.SMTPFrom) != ""
}