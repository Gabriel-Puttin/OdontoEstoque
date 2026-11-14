package models

import (
	"errors"
	"time"
)

type Product struct {
	ID                int64     `json:"id"`
	Nome              string    `json:"nome"`
	Variacao          *string   `json:"variacao"`
	NomeExibicao      string    `json:"nome_exibicao"`
	TipoProdutoCodigo string    `json:"tipo_produto_codigo"`
	TipoProduto       string    `json:"tipo_produto"`
	Grupo             string    `json:"grupo"`
	CategoriaID       int16     `json:"categoria_id"`
	Categoria         string    `json:"categoria"`
	UnidadeMedidaID   int16     `json:"unidade_medida_id"`
	UnidadeMedida     string    `json:"unidade_medida"`
	QuantidadeAtual   int32     `json:"quantidade_atual"`
	EstoqueMinimo     int32     `json:"estoque_minimo"`
	Situacao          string    `json:"situacao"`
	Faltam            int32     `json:"faltam"`
	Validade          *string   `json:"validade"`
	DiasParaVencer    *int32    `json:"dias_para_vencer"`
	CriadoEm          time.Time `json:"criado_em"`
	AtualizadoEm      time.Time `json:"atualizado_em"`
}

// ProductRequest é o corpo do POST e do PUT de produtos.
//   - Quantidade no POST: saldo inicial (padrão 0), registrado como movimentação de 'entrada'.
//   - Quantidade no PUT: novo saldo; se mudar, vira uma movimentação de 'ajuste'. Omitida = mantém.
//   - UsuarioID: responsável pela movimentação (opcional enquanto não há autenticação).
type ProductRequest struct {
	TipoProdutoCodigo string  `json:"tipo_produto_codigo"`
	CategoriaID       int16   `json:"categoria_id"`
	UnidadeMedidaID   int16   `json:"unidade_medida_id"`
	Nome              string  `json:"nome"`
	Variacao          *string `json:"variacao"`
	EstoqueMinimo     *int32  `json:"estoque_minimo"`
	Validade          *string `json:"validade"` // YYYY-MM-DD
	Quantidade        *int32  `json:"quantidade"`
	UsuarioID         *int64  `json:"usuario_id"`
}

// ProductInput são os dados cadastrais já validados que vão para o banco.
type ProductInput struct {
	TipoProdutoCodigo string
	CategoriaID       int16
	UnidadeMedidaID   int16
	Nome              string
	Variacao          *string
	EstoqueMinimo     int32
	Validade          *time.Time
}

var (
	ErrProductNotFound = errors.New("produto não encontrado")
	ErrProductExists   = errors.New("já existe um produto ativo com esse nome e variação")
)

// RuleError é uma regra de negócio recusada pelo banco (ex.: trigger de validade/variação).
type RuleError struct {
	Message string
}

func (e *RuleError) Error() string { return e.Message }
