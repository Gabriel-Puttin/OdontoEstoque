// Package queries expõe os scripts SQL da aplicação.
// Cada query vive num arquivo .sql e é embutida no binário em tempo de compilação,
// então o SQL executado no banco é exatamente o que está nesses arquivos.
package queries

import _ "embed"

// ---------- usuarios ----------

//go:embed usuarios/insert.sql
var UsuarioInsert string

//go:embed usuarios/select_by_email.sql
var UsuarioSelectByEmail string

//go:embed usuarios/update.sql
var UsuarioUpdate string

// ---------- categorias ----------

//go:embed categorias/select_all.sql
var CategoriaSelectAll string

//go:embed categorias/select_by_id.sql
var CategoriaSelectByID string

//go:embed categorias/select_produtos.sql
var CategoriaSelectProdutos string

//go:embed categorias/insert.sql
var CategoriaInsert string

//go:embed categorias/update.sql
var CategoriaUpdate string

//go:embed categorias/delete.sql
var CategoriaDelete string

// ---------- produtos ----------

//go:embed produtos/select_all.sql
var ProdutoSelectAll string

//go:embed produtos/select_by_id.sql
var ProdutoSelectByID string

//go:embed produtos/select_saldo_for_update.sql
var ProdutoSelectSaldoForUpdate string

//go:embed produtos/insert.sql
var ProdutoInsert string

//go:embed produtos/update.sql
var ProdutoUpdate string

//go:embed produtos/delete.sql
var ProdutoDelete string

// ---------- movimentacoes_estoque ----------

//go:embed movimentacoes/insert.sql
var MovimentacaoInsert string

// ---------- idiomas ----------

//go:embed idiomas/select_all.sql
var IdiomaSelectAll string

// ---------- unidades_medida ----------

//go:embed unidades_medida/select_all.sql
var UnidadeMedidaSelectAll string
