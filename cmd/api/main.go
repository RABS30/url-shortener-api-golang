package main

import (
	"log"
	"net/http"
	"os"
	"shorter-url/internal/config"
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

	googleOauthConfig := config.NewGoogleOauthConfig()

	hasher := helper.NewBcryptHasher()

	expireJWT := os.Getenv("JWT_EXPIRE_DURATION")
	expireToken, err := time.ParseDuration(expireJWT)
	if err != nil {
		expireToken = 1 * time.Hour
	}

	cookieConfig := &handler.CookieConfig{
		Domain: os.Getenv("APP_DOMAIN"),
		MaxAge: int(expireToken.Seconds()),
		Secure: os.Getenv("APP_ENV") == "production",
		Path:   "/",
	}

	middleware.InitLogger()

	clickEventRepo := repository.NewClickEventsRepository(database)
	clickEventService := service.NewClickEventService(clickEventRepo)
	clickEventHandler := handler.NewClickEventHandler(clickEventService)

	shortUrlRepo := repository.NewShortUrlRepository(database)
	shortUrlService := service.NewShortUrlService(shortUrlRepo)
	shortUrlHandler := handler.NewShortUrlHandler(shortUrlService, clickEventService)

	userRepo := repository.NewUserRepository(database)

	userOtpsRepo := repository.NewUserOtpsRepository(database)
	userOtpsService := service.NewUserOtpsService(userOtpsRepo, emailService, userRepo, []byte(JwtSecret))
	userOtpsHandler := handler.NewUserOtpsHandler(userOtpsService, []byte(JwtSecret))

	userService := service.NewUserService(userRepo, []byte(JwtSecret), hasher, database, userOtpsService)
	userHandler := handler.NewUserHandler(userService, userOtpsService, cookieConfig, []byte(JwtSecret))

	oauthGoogleHandler := handler.NewOauthGoogleHandler(userService, googleOauthConfig, *cookieConfig)

	authMiddleware := middleware.NewAuthMiddleware(userRepo, []byte(JwtSecret))

	router := httprouter.New()

	router.GET("/r/:shortCode", shortUrlHandler.AccessShortCode)

	router.GET("/", userHandler.HelloWorld)
	router.POST("/user/login", middleware.GuestOnly(JwtSecret)(userHandler.Login))
	router.POST("/user/register", middleware.GuestOnly(JwtSecret)(userHandler.Register))
	router.POST("/user/reset-password", userHandler.ResetPassword)
	router.POST("/user/change-password", userHandler.ChangePassword)
	router.GET("/user/verify", authMiddleware.Authenticate(authMiddleware.VerifiedOnly(userHandler.VerifyUser)))

	router.GET("/user/login/google", oauthGoogleHandler.Login)
	router.GET("/auth/google/callback", oauthGoogleHandler.Callback)

	// router.POST("/send-otp", userOtpsHandler.RequestOTP)
	router.POST("/verify-otp", userOtpsHandler.VerifyOTP)
	router.GET("/verify-session-otp", userOtpsHandler.VerifySessionOtpPage)

	// router.POST("/api/urls", authMiddleware.Authenticate(JwtSecret)(middleware.VerifiedUserOnly(shortUrlHandler.Create)))
	// router.GET("/api/urls/:shortUrlId/analytics", middleware.Authenticate(JwtSecret)(middleware.VerifiedUserOnly(clickEventHandler.FindByShortUrlId)))

	router.POST("/api/urls", authMiddleware.Authenticate(authMiddleware.VerifiedOnly(shortUrlHandler.Create)))
	router.GET("/api/urls/:shortUrlId/analytics", authMiddleware.Authenticate(authMiddleware.VerifiedOnly(clickEventHandler.FindByShortUrlId)))

	logger := middleware.Logger(router)
	requestID := middleware.RequestID(logger)

	log.Println("Server running on port :8080")

	server := http.Server{
		Addr:    os.Getenv("APP_ADDR"),
		Handler: requestID,
	}

	log.Fatal(server.ListenAndServe())
}
