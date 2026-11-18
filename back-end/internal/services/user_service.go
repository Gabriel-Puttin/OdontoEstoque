package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"odontoestoque/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, nome, email, senhaHash string, idioma *string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id int64, nome, email, idioma string) (*models.User, error)
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("senha-inexistente"), bcrypt.DefaultCost)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
	nome := strings.TrimSpace(req.Nome)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	fields := map[string]string{}
	validateNomeEmail(nome, email, fields)
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

// Update altera nome, e-mail e idioma. A senha não é alterada aqui.
func (s *UserService) Update(ctx context.Context, id int64, req models.UpdateUserRequest) (*models.User, error) {
	nome := strings.TrimSpace(req.Nome)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	idioma := strings.TrimSpace(req.IdiomaCodigo)

	fields := map[string]string{}
	validateNomeEmail(nome, email, fields)
	if idioma == "" {
		fields["idioma_codigo"] = "obrigatório"
	}
	if len(fields) > 0 {
		return nil, &models.ValidationError{Fields: fields}
	}

	return s.repo.Update(ctx, id, nome, email, idioma)
}

// validateNomeEmail aplica as regras de nome e e-mail usadas no cadastro e na edição.
func validateNomeEmail(nome, email string, fields map[string]string) {
	if nome == "" {
		fields["nome"] = "obrigatório"
	} else if utf8.RuneCountInString(nome) > 120 {
		fields["nome"] = "máximo de 120 caracteres"
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email || len(email) > 254 {
		fields["email"] = "e-mail inválido"
	}
}

func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (*models.User, error) {
	email := strings.TrimSpace(req.Email)
	if email == "" || req.Senha == "" {
		return nil, models.ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, models.ErrUserNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Senha))
		return nil, models.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.SenhaHash), []byte(req.Senha)); err != nil {
		return nil, models.ErrInvalidCredentials
	}
	return user, nil
}
