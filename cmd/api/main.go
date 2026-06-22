package main

import (
	"log"
	"net/http"
	"os"
	"shorter-url/internal/database"
	"shorter-url/internal/handler"
	"shorter-url/internal/middleware"
	"shorter-url/internal/repository"
	"shorter-url/internal/service"

	"github.com/julienschmidt/httprouter"
)

func main() {
	database := database.DatabaseConnect()
	defer database.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("This is jwt token : ", jwtSecret)
		jwtSecret = "super-secret-key-fallback-12345"
		log.Println("Warning: JWT_SECRET env is not set, using default fallback key")
	}

	shortUrlRepo := repository.NewShortUrlRepository(database)
	shortUrlService := service.NewShortUrlService(shortUrlRepo)
	shortUrlHandler := handler.NewShortUrlHandler(shortUrlService)

	userRepo := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepo, []byte(jwtSecret))
	userHandler := handler.NewUserHandler(userService)

	router := httprouter.New()

	router.POST("/api/urls", middleware.AuthMiddleware(jwtSecret)(shortUrlHandler.Create))
	router.GET("/:shortCode", shortUrlHandler.AccessShortCode)

	router.POST("/user/login", userHandler.Login)
	router.POST("/user/register", userHandler.Register)

	logger := middleware.Logger(router)

	log.Println("Server running on port :8080")

	server := http.Server{
		Addr:    ":8080",
		Handler: logger,
	}

	log.Fatal(server.ListenAndServe())
}
