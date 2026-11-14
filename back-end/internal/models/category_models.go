package models

import (
	"errors"
	"time"
)

type Category struct {
	ID           int16     `json:"id"`
	Nome         string    `json:"nome"`
	Itens        *int64    `json:"itens,omitempty"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

// CategoryRequest é o corpo do POST e do PUT de categorias.
type CategoryRequest struct {
	Nome string `json:"nome"`
}

// CategoryProduct é um produto listado dentro de uma categoria.
type CategoryProduct struct {
	ID                int64   `json:"id"`
	Nome              string  `json:"nome"`
	Variacao          *string `json:"variacao"`
	NomeExibicao      string  `json:"nome_exibicao"`
	TipoProdutoCodigo string  `json:"tipo_produto_codigo"`
	TipoProduto       string  `json:"tipo_produto"`
	Grupo             string  `json:"grupo"`
	UnidadeMedida     string  `json:"unidade_medida"`
	QuantidadeAtual   int32   `json:"quantidade_atual"`
	EstoqueMinimo     int32   `json:"estoque_minimo"`
	Validade          *string `json:"validade"`
}

type CategoryProductsResponse struct {
	Categoria Category          `json:"categoria"`
	Produtos  []CategoryProduct `json:"produtos"`
}

var (
	ErrCategoryNotFound    = errors.New("categoria não encontrada")
	ErrCategoryNameExists  = errors.New("já existe uma categoria com esse nome")
	ErrCategoryHasProducts = errors.New("a categoria possui produtos vinculados e não pode ser excluída")
)
