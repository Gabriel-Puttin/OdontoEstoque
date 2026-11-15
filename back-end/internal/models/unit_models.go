package models

import "time"

type Unit struct {
	ID       int16     `json:"id"`
	Nome     string    `json:"nome"`
	CriadoEm time.Time `json:"criado_em"`
}
