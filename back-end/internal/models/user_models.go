package models

import (
	"errors"
	"time"
)

// User representa uma linha da tabela usuarios. SenhaHash nunca vai para o JSON.
type User struct {
	ID           int64     `json:"id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email"`
	SenhaHash    string    `json:"-"`
	IdiomaCodigo string    `json:"idioma_codigo"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

// CreateUserRequest é o corpo do POST /api/v1/users.
type CreateUserRequest struct {
	Nome           string  `json:"nome"`
	Email          string  `json:"email"`
	Senha          string  `json:"senha"`
	ConfirmarSenha string  `json:"confirmar_senha"`
	IdiomaCodigo   *string `json:"idioma_codigo,omitempty"`
}

// Erros de domínio — o handler traduz cada um para um status HTTP.
var (
	ErrEmailAlreadyExists = errors.New("e-mail já cadastrado")
	ErrInvalidLanguage    = errors.New("idioma inválido")
)

// ValidationError carrega mensagens de validação por campo.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "dados inválidos" }
