package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"odontoestoque/config"
	"odontoestoque/infra/database"
	"odontoestoque/internal/handlers"
	"odontoestoque/internal/repositories"
	"odontoestoque/internal/services"
	"odontoestoque/pkg/logger"
	"odontoestoque/internal/middleware"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ERRO AO CARREGAR CONFIGURAÇÃO: %v", err)
	}

	appLogger := logger.New(cfg.Env)
	appLogger.Info("OdontoEstoque API iniciando...", "port", cfg.Server.Port)

	// PostgreSQL
	db, err := database.Connect(context.Background(), cfg.Database.DSN, appLogger)
	if err != nil {
		log.Fatalf("PostgreSQL indisponível: %v", err)
	}
	defer db.Close()

	// Repositórios (adaptadores de saída)
	userRepo := repositories.NewUserRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	productRepo := repositories.NewProductRepository(db)
	languageRepo := repositories.NewLanguageRepository(db)
	unitRepo := repositories.NewUnitRepository(db)

	// Services (núcleo / regras de negócio)
	userService := services.NewUserService(userRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	productService := services.NewProductService(productRepo)
	languageService := services.NewLanguageService(languageRepo)
	unitService := services.NewUnitService(unitRepo)

	// Handlers (adaptadores de entrada)
	userHandler := handlers.NewUserHandler(userService, appLogger)
	categoryHandler := handlers.NewCategoryHandler(categoryService, appLogger)
	productHandler := handlers.NewProductHandler(productService, appLogger)
	languageHandler := handlers.NewLanguageHandler(languageService, appLogger)
	unitHandler := handlers.NewUnitHandler(unitService, appLogger)

	// Roteador
	mux := http.NewServeMux()

	// Usuários
	mux.HandleFunc("POST /api/users", userHandler.Create)
	mux.HandleFunc("POST /api/users/login", userHandler.Login)
	mux.HandleFunc("POST /api/users/logout", userHandler.Logout)

	// Categorias
	mux.HandleFunc("GET /api/categories", categoryHandler.List)
	mux.HandleFunc("POST /api/categories", categoryHandler.Create)
	mux.HandleFunc("GET /api/categories/{id}/products", categoryHandler.Products)
	mux.HandleFunc("PUT /api/categories/{id}", categoryHandler.Update)
	mux.HandleFunc("DELETE /api/categories/{id}", categoryHandler.Delete)

	// Produtos
	mux.HandleFunc("GET /api/products", productHandler.List)
	mux.HandleFunc("POST /api/products", productHandler.Create)
	mux.HandleFunc("GET /api/products/{id}", productHandler.Get)
	mux.HandleFunc("PUT /api/products/{id}", productHandler.Update)
	mux.HandleFunc("DELETE /api/products/{id}", productHandler.Delete)

	// Idiomas e unidades de medida (somente leitura)
	mux.HandleFunc("GET /api/languages", languageHandler.List)
	mux.HandleFunc("GET /api/units", unitHandler.List)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	appLogger.Info("servidor HTTP iniciado", "addr", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: middleware.CORS(mux),
	}

	if err := server.ListenAndServe(); err != nil {
		appLogger.Error("servidor encerrado", "error", err)
		os.Exit(1)
	}
}
