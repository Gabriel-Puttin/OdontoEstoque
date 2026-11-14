// Package queries expõe os scripts SQL da aplicação.
// Cada query vive num arquivo .sql e é embutida no binário em tempo de compilação,
// então o SQL executado no banco é exatamente o que está nesses arquivos.
package queries

import _ "embed"

// ---------- usuarios ----------

//go:embed usuarios/insert.sql
var UsuarioInsert string
