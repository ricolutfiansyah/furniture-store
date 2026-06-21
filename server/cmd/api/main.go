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

	db := database.ConnectDB(cfg.DBUrl)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)

	authService := service.NewUserService(userRepo, cfg.JWTSecret)
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

	r.Route("api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.login)
		r.Get("/product", productHandler.GetAll)
		r.Get("/product/{slug}", productHandler.GetBySlug)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("users/me", authHandler.GetProfile)
		r.Get("users/me", authHandler.UpdateProfile)

		r.Post("/orders", orderHandler.Checkout)
		r.Get("/orders", orderHandler.GetOrders)

		r.Post("/wishlists", wishlistHandler.Add)
		r.Get("/wishlists", wishlistHandler.GetAll)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.AdminOnly)

		r.Get("/admin/orders", adminHandler.GetOrders)
		r.Put("/admin/orders/{id}/confirm-payment", adminHandler.ConfirmPayment)
		r.Put("/admin/orders/{id}/status", adminHandler.UpdateStatus)

		r.Post("/admin/products", adminHandler.CreateProduct)
		r.Put("/admin/products/{id}", adminHandler.UpdateProduct)
		r.Delete("/admin/products/{id}", adminHandler.DeleteProduct)
	})

	fmt.Printf("Starting server on port %s...\n", cfg.Port)
	addr := fmt.Sprintf(":%s", cfg.Port)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
