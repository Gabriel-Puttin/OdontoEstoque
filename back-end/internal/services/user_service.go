package services

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"odontoestoque/internal/models"
)

// UserRepository é a porta (port) que o serviço precisa. Qualquer adaptador
// que implemente esse contrato (Postgres, mock de teste…) pode ser plugado.
type UserRepository interface {
	Create(ctx context.Context, nome, email, senhaHash string, idioma *string) (*models.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create valida os dados, gera o hash da senha e persiste o usuário.
func (s *UserService) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	nome := strings.TrimSpace(req.Nome)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	fields := map[string]string{}
	if nome == "" {
		fields["nome"] = "obrigatório"
	} else if utf8.RuneCountInString(nome) > 120 {
		fields["nome"] = "máximo de 120 caracteres"
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email || len(email) > 254 {
		fields["email"] = "e-mail inválido"
	}
	if utf8.RuneCountInString(req.Senha) < 8 {
		fields["senha"] = "mínimo de 8 caracteres"
	} else if len(req.Senha) > 72 {
		fields["senha"] = "máximo de 72 bytes"
	}
	if req.Senha != req.ConfirmarSenha {
		fields["confirmar_senha"] = "as senhas não coincidem"
	}
	if len(fields) > 0 {
		return nil, &models.ValidationError{Fields: fields}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("user service: hash da senha: %w", err)
	}

	return s.repo.Create(ctx, nome, email, string(hash), req.IdiomaCodigo)
}
