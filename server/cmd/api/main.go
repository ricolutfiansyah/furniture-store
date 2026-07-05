package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"furniture-api/internal/config"
	"furniture-api/internal/handler"
	"furniture-api/internal/infrastructure/database"
	"furniture-api/internal/middleware"
	"furniture-api/internal/repository"
	"furniture-api/internal/response"
	"furniture-api/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	db, err := database.ConnectDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	// --- Repository ---
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// --- Service ---
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	productService := service.NewProductService(productRepo)
	cartService := service.NewCartService(cartRepo, productRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, db)

	// --- Handler ---
	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)

	r := chi.NewRouter()

	// CORS options
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// server check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			response.WriteError(w, http.StatusServiceUnavailable, "database unavailale")
			return
		}
		response.WriteSuccess(w, http.StatusOK, nil, "server is healthy")
	})

	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret, userRepo)

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Get("/products", productHandler.GetAllProduct)
		r.Get("/products/{slug}", productHandler.GetProductBySlug)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			r.Get("/users/me", authHandler.GetProfile)

			r.Get("/cart", cartHandler.GetCart)
			r.Post("/cart", cartHandler.AddToCart)
			r.Put("/cart/items/{id}", cartHandler.UpdateQuantity)
			r.Delete("/cart/items/{id}", cartHandler.RemoveItem)

			r.Post("/orders/checkout", orderHandler.Checkout)
			r.Get("/orders", orderHandler.GetOrders)
			r.Get("/orders/{id}", orderHandler.GetOrderDetail)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Use(middleware.AdminMiddleware)

			r.Patch("orders/{id}/status", orderHandler.UpdateStatus)
		})
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("starting server on port %s...", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed to start: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}
