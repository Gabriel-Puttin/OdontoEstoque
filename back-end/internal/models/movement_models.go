package models

import "time"

// Movement é uma linha do histórico (movimentacoes_estoque + produto, unidade e usuário).
type Movement struct {
	ID                  int64     `json:"id"`
	Tipo                string    `json:"tipo"`
	ProdutoID           int64     `json:"produto_id"`
	Produto             string    `json:"produto"`
	ProdutoExcluido     bool      `json:"produto_excluido"`
	QuantidadeAnterior  int32     `json:"quantidade_anterior"`
	QuantidadePosterior int32     `json:"quantidade_posterior"`
	Delta               int32     `json:"delta"`
	UnidadeMedida       string    `json:"unidade_medida"`
	UsuarioID           *int64    `json:"usuario_id"`
	Usuario             *string   `json:"usuario"`
	CriadoEm            time.Time `json:"criado_em"`
}

// MovementFilter são os filtros do GET /api/movements.
type MovementFilter struct {
	Tipo   *string // nil = todos
	Limite int
}
