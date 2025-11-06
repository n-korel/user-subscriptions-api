package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/n-korel/user-subscriptions-api/docs" // swagger docs
	"github.com/n-korel/user-subscriptions-api/internal/logger"
	"github.com/n-korel/user-subscriptions-api/internal/subscriptions"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

//	@title			User Subscriptions API
//	@version		1.0
//	@description	REST API для управления подписками пользователей
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	https://github.com/n-korel
//	@contact.email	example@mail.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

// @host		localhost:8080
// @BasePath	/v1
func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		fmt.Println("FATAL: DSN environment variable is not set")
		os.Exit(1)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	log, err := logger.New(logLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Failed to connect to database", map[string]interface{}{"error": err})
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping database", map[string]interface{}{"error": err})
	}

	log.Info("Database has connected!", nil)

	repo := subscriptions.NewRepository(db, log)
	service := subscriptions.NewService(repo, log)
	handler := subscriptions.NewHandler(service, log)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	handler.RegisterRoutes(r)

	// Swagger endpoint
	r.Route("/v1/swagger", func(r chi.Router) {
		r.Handle("/*", httpSwagger.Handler())
	})

	log.Info("Server starting", map[string]interface{}{"port": port})
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Server error", map[string]interface{}{"error": err})
	}
}
