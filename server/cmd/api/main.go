package main

import (
	"fmt"
	"log"
	"net/http"

	"furniture-api/internal/config"
	"furniture-api/internal/handler"
	"furniture-api/internal/infrastructure/database"
	"furniture-api/internal/middleware"
	"furniture-api/internal/repository"
	"furniture-api/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	db, err := database.ConnectDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	productService := service.NewProductService(productRepo)

	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is healthy and running"))
	})

	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Get("/products", productHandler.GetAllProduct)
		r.Get("/products/{slug}", productHandler.GetProductBySlug)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Get("/users/me", authHandler.GetProfile)
		})
	})

	fmt.Printf("Starting server on port %s...\n", cfg.Port)
	addr := fmt.Sprintf(":%s", cfg.Port)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
