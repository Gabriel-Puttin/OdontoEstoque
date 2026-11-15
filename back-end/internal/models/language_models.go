package models

import "time"

type Language struct {
	Codigo   string    `json:"codigo"`
	Nome     string    `json:"nome"`
	Regiao   string    `json:"regiao"`
	Ativo    bool      `json:"ativo"`
	CriadoEm time.Time `json:"criado_em"`
}
