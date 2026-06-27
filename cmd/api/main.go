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

	"github.com/julienschmidt/httprouter"
)

func main() {
	database := database.DatabaseConnect()
	defer database.Close()

	JwtSecret := os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		log.Println("Warning: JWT_SECRET env is not set, using default fallback key")
	}

	emailSender := helper.NewEmailSender("smtp.gmail.com", "587", "emailforhostuser@gmail.com", "fipdijyxekwufmlp")
	hasher := helper.NewBcryptHasher()
	baseUrl := os.Getenv("APP_HOST")

	clickEventRepo := repository.NewClickEventsRepository(database)
	clickEventService := service.NewClickEventService(clickEventRepo)
	clickEventHandler := handler.NewClickEventHandler(clickEventService)

	shortUrlRepo := repository.NewShortUrlRepository(database)
	shortUrlService := service.NewShortUrlService(shortUrlRepo)
	shortUrlHandler := handler.NewShortUrlHandler(shortUrlService, clickEventService)

	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo, []byte(JwtSecret))
	userHandler := handler.NewUserHandler(userService)

	passwordResetRepo := repository.NewPasswordResetTokensRepository(database)
	passwordResetService := service.NewPasswordResetTokensService(passwordResetRepo, userRepo, emailSender, hasher, baseUrl)
	passwordResethandler := handler.NewPasswordResetTokensHandler(passwordResetService)

	router := httprouter.New()

	router.GET("/r/:shortCode", shortUrlHandler.AccessShortCode)

	router.POST("/user/login", userHandler.Login)
	router.POST("/user/register", userHandler.Register)

	router.POST("/forgot-password", passwordResethandler.ForgotPasswordHandler)
	router.POST("/reset-password", passwordResethandler.ResetPasswordHandler)

	router.POST("/api/urls", middleware.AuthMiddleware(JwtSecret)(shortUrlHandler.Create))
	router.GET("/api/urls/:shortUrlId/analytics", middleware.AuthMiddleware(JwtSecret)(clickEventHandler.FindByShortUrlId))

	logger := middleware.Logger(router)

	log.Println("Server running on port :8080")

	server := http.Server{
		Addr:    ":8080",
		Handler: logger,
	}

	log.Fatal(server.ListenAndServe())
}
