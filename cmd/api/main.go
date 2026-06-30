package main

import (
	"log"
	"net/http"
	"os"
	"shorter-url/internal/database"
	"shorter-url/internal/handler"
	"shorter-url/internal/helper"
	"shorter-url/internal/middleware"
	"shorter-url/internal/repository"
	"shorter-url/internal/service"
	"time"

	"github.com/julienschmidt/httprouter"
)

func main() {
	database := database.DatabaseConnect()
	defer database.Close()

	JwtSecret := os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		log.Fatal("Warning: JWT_SECRET env is not set")
	}

	emailService := helper.NewEmailService(os.Getenv("MAIL_HOST"), os.Getenv("MAIL_PORT"), os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"))

	hasher := helper.NewBcryptHasher()

	baseUrl := os.Getenv("APP_HOST")

	expireJWT := os.Getenv("JWT_EXPIRE_DURATION")
	expireToken, err := time.ParseDuration(expireJWT)
	if err != nil {
		expireToken = 1 * time.Hour
	}

	cookieConfig := &handler.CookieConfig{
		Domain: os.Getenv("APP_DOMAIN"),
		MaxAge: int(expireToken.Seconds()),
		Secure: os.Getenv("APP_ENV") == "production",
	}

	clickEventRepo := repository.NewClickEventsRepository(database)
	clickEventService := service.NewClickEventService(clickEventRepo)
	clickEventHandler := handler.NewClickEventHandler(clickEventService)

	shortUrlRepo := repository.NewShortUrlRepository(database)
	shortUrlService := service.NewShortUrlService(shortUrlRepo)
	shortUrlHandler := handler.NewShortUrlHandler(shortUrlService, clickEventService)

	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo, []byte(JwtSecret))
	userHandler := handler.NewUserHandler(userService, cookieConfig)

	passwordResetRepo := repository.NewPasswordResetTokensRepository(database)
	passwordResetService := service.NewPasswordResetTokensService(passwordResetRepo, userRepo, emailService, hasher, baseUrl)
	passwordResethandler := handler.NewPasswordResetTokensHandler(passwordResetService)

	verificationRepo := repository.NewVerificationTokenRepository(database)
	verificationService := service.NewVerificationTokenService(verificationRepo, userRepo, emailService, baseUrl)
	verificationHandler := handler.NewVerificationTokenHandler(verificationService)

	router := httprouter.New()

	router.GET("/r/:shortCode", shortUrlHandler.AccessShortCode)

	router.POST("/user/login", middleware.GuestOnly(JwtSecret)(userHandler.Login))
	router.POST("/user/register", middleware.GuestOnly(JwtSecret)(userHandler.Register))

	router.POST("/user/verify", verificationHandler.RequestVerification)
	router.GET("/verify", verificationHandler.VerificationAccount)

	router.POST("/forgot-password", passwordResethandler.ForgotPasswordHandler)
	router.POST("/reset-password", passwordResethandler.ResetPasswordHandler)

	router.POST("/api/urls", middleware.AuthMiddleware(JwtSecret)(middleware.VerifiedUserOnly(shortUrlHandler.Create)))
	router.GET("/api/urls/:shortUrlId/analytics", middleware.AuthMiddleware(JwtSecret)(middleware.VerifiedUserOnly(clickEventHandler.FindByShortUrlId)))

	logger := middleware.Logger(router)

	log.Println("Server running on port :8080")

	server := http.Server{
		Addr:    os.Getenv("APP_ADDR"),
		Handler: logger,
	}

	log.Fatal(server.ListenAndServe())
}
