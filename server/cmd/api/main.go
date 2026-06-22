package main

import (
	"fmt"
	"log"
	"net/http"

	"furniture-api/internal/config"
	"furniture-api/internal/handler"
	"furniture-api/internal/infrastructure/database"
	"furniture-api/internal/repository"
	"furniture-api/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	db, _ := database.ConnectDB(cfg.DBUrl)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)

	authService := service.NewAuthService(userRepo)

	authHandler := handler.NewAuthHandler(authService)

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

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
	})

	fmt.Printf("Starting server on port %s...\n", cfg.Port)
	addr := fmt.Sprintf(":%s", cfg.Port)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
