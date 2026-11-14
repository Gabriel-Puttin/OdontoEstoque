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

	// Services (núcleo / regras de negócio)
	userService := services.NewUserService(userRepo)

	// Handlers (adaptadores de entrada)
	userHandler := handlers.NewUserHandler(userService, appLogger)

	// Roteador
	mux := http.NewServeMux()

	// Usuários
	mux.HandleFunc("POST /api/users", userHandler.Create)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	appLogger.Info("servidor HTTP iniciado", "addr", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		appLogger.Error("servidor encerrado", "error", err)
		os.Exit(1)
	}
}
