# Modelagem de dados: OdontoEstoque

## Etapa 1: Levantamento de entidades

| Entidade | Responsabilidade | Atributos principais |
| :--- | :--- | :--- |
| `idiomas` | Idiomas do app (tela Idioma) | `codigo` PK (`pt-BR`), `nome`, `regiao`, `ativo` |
| `usuarios` | Cadastro, login e perfil | `id`, `nome`, `email` (único), `senha_hash`, `idioma_codigo` FK |
| `categorias` | Categorias de produtos | `id`, `nome` (único) |
| `unidades_medida` | caixa, unidade, tubo, pacote | `id`, `nome` (único) |
| `tipos_produto` | Regras de validade e variação | `codigo` PK, `nome`, `descricao`, `grupo`, `exige_validade`, `tipo_variacao` |
| `produtos` | Itens de estoque | `id`, `tipo`/`categoria`/`unidade` FK, `nome`, `variacao`, `quantidade_atual`, `estoque_minimo`, `validade`, `excluido_em` |
| `movimentacoes_estoque` | Histórico permanente | `id`, `produto_id`, `usuario_id`, `tipo`, `quantidade_anterior`, `quantidade_posterior`, `delta` (gerado), `unidade_medida_id` |

---

## Etapa 2: DER

![Diagrama Entidade-Relacionamento](./Diagrama_ER.png)

*(Substitua o caminho acima pela URL ou diretório onde a imagem do DER será salva no repositório).*

---

## Etapa 3: Dicionário de dados (resumo)

| Tabela.campo | Regra |
| :--- | :--- |
| `usuarios.email` | Formato validado e único sem diferenciar maiúsculas (índice em `lower(email)`). |
| `usuarios.senha_hash` | Só o hash (`bcrypt`/`argon2id`), com `CHECK` de tamanho mínimo. O mínimo de 8 caracteres e a confirmação de senha são validados na aplicação. |
| `usuarios.idioma_codigo` | FK para `idiomas`, padrão `pt-BR`. |
| `categorias.nome` | Único sem diferenciar maiúsculas. FK de `produtos` com `RESTRICT`, então não se apaga categoria em uso. |
| `unidades_medida.nome` | Singular e único. A pluralização (“12 caixas”) fica no i18n. |
| `tipos_produto.grupo` / `exige_validade` | `CHECK`: perecível se e somente se exige validade. |
| `tipos_produto.tipo_variacao` | `cor`, `tamanho` ou `NULL` (sem variação). |
| `produtos.nome` / `variacao` | Nome exibido = `nome` + `variacao` (“Resina composta A2”). Índice único parcial entre itens ativos. |
| `produtos.quantidade_atual` | Inteiro $\ge 0$. Nasce com `0` e só muda por movimentação. |
| `produtos.estoque_minimo` | Inteiro $\ge 0$. |
| `produtos.validade` | Obrigatória ou proibida conforme o tipo (trigger). |
| `produtos.excluido_em` | `NULL` significa ativo. |
| `movimentacoes_estoque.tipo` | `entrada`, `saida` ou `ajuste`. |
| `movimentacoes_estoque.quantidade_anterior` / `quantidade_posterior` | Reproduzem o “12 → 11” da tela. O trigger rejeita se a anterior não bater com o saldo atual (proteção contra concorrência). |
| `movimentacoes_estoque.delta` | Gerada: `posterior - anterior`. |
| `movimentacoes_estoque.usuario_id` | Anulável, `SET NULL` se o usuário for removido. Inserir um item novo com saldo 3 vira uma entrada `0 → 3`, como no histórico do protótipo. |
---

## Etapa 4: Documentação da API

API REST em Go (arquitetura hexagonal, SQL puro com `pgx`, sem ORM).

### Como executar

```bash
cd back-end
go run ./cmd
```

| Variável (`back-end/.env`) | Obrigatória | Padrão | Descrição |
| :--- | :--- | :--- | :--- |
| `CONNECTION_STRING` | Sim | — | String de conexão do PostgreSQL |
| `PORT` | Não | `8080` | Porta do servidor HTTP |
| `APP_ENV` | Não | `development` | `production` gera logs em JSON |

URL base: `http://localhost:8080`

### Convenções

- Corpo das requisições e respostas em JSON (`Content-Type: application/json`).
- Datas sem hora no formato `AAAA-MM-DD`; data e hora em ISO 8601 com fuso.
- Campos desconhecidos no corpo são rejeitados com `400` (exceto em `/api/users/login`).
- Não há autenticação nem token neste protótipo.
- Formato padrão de erro:

```json
{ "error": "mensagem do erro" }
```

Erros de validação trazem também o detalhe por campo:

```json
{
  "error": "dados inválidos",
  "fields": {
    "nome": "obrigatório",
    "estoque_minimo": "não pode ser negativo"
  }
}
```

| Status | Quando |
| :--- | :--- |
| `200` | Sucesso |
| `201` | Recurso criado |
| `204` | Sucesso sem corpo (exclusões) |
| `400` | JSON malformado ou `id` inválido na URL |
| `403` | Login com e-mail ou senha inválidos |
| `404` | Recurso não encontrado |
| `409` | Conflito (nome/e-mail duplicado, categoria em uso) |
| `422` | Dados inválidos ou regra de negócio recusada |
| `500` | Erro interno |

### Resumo das rotas

| Método | Rota | Descrição |
| :--- | :--- | :--- |
| `POST` | `/api/users` | Cadastrar usuário |
| `POST` | `/api/users/login` | Login |
| `POST` | `/api/users/logout` | Logout |
| `PUT` | `/api/users/{id}` | Atualizar usuário |
| `GET` | `/api/categories` | Listar categorias |
| `POST` | `/api/categories` | Cadastrar categoria |
| `GET` | `/api/categories/{id}/products` | Listar produtos de uma categoria |
| `PUT` | `/api/categories/{id}` | Renomear categoria |
| `DELETE` | `/api/categories/{id}` | Excluir categoria |
| `GET` | `/api/products` | Listar produtos |
| `POST` | `/api/products` | Cadastrar produto |
| `GET` | `/api/products/{id}` | Buscar produto |
| `PUT` | `/api/products/{id}` | Editar produto |
| `DELETE` | `/api/products/{id}` | Excluir produto (lógico) |
| `GET` | `/api/languages` | Listar idiomas |
| `GET` | `/api/units` | Listar unidades de medida |
| `GET` | `/api/movements` | Histórico de movimentações |

---

### Usuários

#### `POST /api/users` — Cadastrar usuário

| Campo | Tipo | Obrigatório | Regra |
| :--- | :--- | :--- | :--- |
| `nome` | string | Sim | Até 120 caracteres |
| `email` | string | Sim | E-mail válido, único (sem diferenciar maiúsculas). Salvo em minúsculas |
| `senha` | string | Sim | De 8 caracteres a 72 bytes |
| `confirmar_senha` | string | Sim | Igual a `senha` |
| `idioma_codigo` | string | Não | Código de um idioma existente. Padrão `pt-BR` |

Requisição:

```json
{
  "nome": "Marina Costa",
  "email": "marina.costa@clinicasorriso.com",
  "senha": "Dente@2026",
  "confirmar_senha": "Dente@2026",
  "idioma_codigo": "pt-BR"
}
```

Resposta `201`:

```json
{
  "id": 42,
  "nome": "Marina Costa",
  "email": "marina.costa@clinicasorriso.com",
  "idioma_codigo": "pt-BR",
  "criado_em": "2026-10-03T09:15:22.481207-03:00",
  "atualizado_em": "2026-10-03T09:15:22.481207-03:00"
}
```

Erros: `409` e-mail já cadastrado · `422` campos inválidos ou `idioma_codigo` inexistente.

```json
{
  "error": "dados inválidos",
  "fields": {
    "senha": "mínimo de 8 caracteres",
    "confirmar_senha": "as senhas não coincidem"
  }
}
```

#### `POST /api/users/login` — Login

Não gera token: apenas confirma as credenciais e devolve os dados básicos do usuário. Guarde o `id` no front para usar em `PUT /api/users/{id}`.

Requisição:

```json
{
  "email": "marina.costa@clinicasorriso.com",
  "senha": "Dente@2026"
}
```

Resposta `200`:

```json
{
  "id": 42,
  "nome": "Marina Costa",
  "email": "marina.costa@clinicasorriso.com",
  "idioma_codigo": "pt-BR"
}
```

Resposta `403` (e-mail inexistente ou senha errada — a mensagem é a mesma nos dois casos):

```json
{ "error": "E-mail ou senha inválidos. Confira seus dados e tente novamente." }
```

#### `POST /api/users/logout` — Logout

Sem corpo. Como não há sessão no servidor, apenas confirma; o front descarta os dados do usuário.

Resposta `200`:

```json
{ "message": "logout realizado" }
```

#### `PUT /api/users/{id}` — Atualizar usuário

Altera nome, e-mail e idioma. A senha não muda por esta rota (enviar `senha` retorna `400`). Todos os campos são obrigatórios.

| Campo | Tipo | Obrigatório | Regra |
| :--- | :--- | :--- | :--- |
| `nome` | string | Sim | Até 120 caracteres |
| `email` | string | Sim | E-mail válido, único (sem diferenciar maiúsculas). Salvo em minúsculas |
| `idioma_codigo` | string | Sim | Código de um idioma existente |

Requisição:

```json
{
  "nome": "Marina Costa Ribeiro",
  "email": "marina.ribeiro@clinicasorriso.com",
  "idioma_codigo": "es-ES"
}
```

Resposta `200`:

```json
{
  "id": 42,
  "nome": "Marina Costa Ribeiro",
  "email": "marina.ribeiro@clinicasorriso.com",
  "idioma_codigo": "es-ES",
  "criado_em": "2026-10-03T09:15:22.481207-03:00",
  "atualizado_em": "2026-10-20T18:02:47.330415-03:00"
}
```

Erros: `400` id inválido · `404` usuário não encontrado · `409` e-mail já cadastrado por outro usuário · `422` campos inválidos ou `idioma_codigo` inexistente.

---

### Categorias

#### `GET /api/categories` — Listar categorias

Ordenadas por nome. `itens` é a quantidade de produtos ativos na categoria.

Resposta `200`:

```json
[
  {
    "id": 12,
    "nome": "Endodontia",
    "itens": 7,
    "criado_em": "2026-09-20T14:02:10.118034-03:00",
    "atualizado_em": "2026-09-20T14:02:10.118034-03:00"
  },
  {
    "id": 15,
    "nome": "Ortodontia",
    "itens": 0,
    "criado_em": "2026-09-21T08:44:51.902311-03:00",
    "atualizado_em": "2026-09-28T17:30:05.774120-03:00"
  }
]
```

#### `POST /api/categories` — Cadastrar categoria

| Campo | Tipo | Obrigatório | Regra |
| :--- | :--- | :--- | :--- |
| `nome` | string | Sim | Até 80 caracteres, único (sem diferenciar maiúsculas) |

Requisição:

```json
{ "nome": "Periodontia" }
```

Resposta `201`:

```json
{
  "id": 18,
  "nome": "Periodontia",
  "criado_em": "2026-10-05T10:11:42.335901-03:00",
  "atualizado_em": "2026-10-05T10:11:42.335901-03:00"
}
```

Erros: `409` nome já existe · `422` nome vazio ou longo demais.

#### `GET /api/categories/{id}/products` — Produtos da categoria

Devolve a categoria e seus produtos ativos (join com tipo e unidade de medida).

Resposta `200`:

```json
{
  "categoria": {
    "id": 12,
    "nome": "Endodontia",
    "criado_em": "2026-09-20T14:02:10.118034-03:00",
    "atualizado_em": "2026-09-20T14:02:10.118034-03:00"
  },
  "produtos": [
    {
      "id": 301,
      "nome": "Lima endodôntica",
      "variacao": "25 mm",
      "nome_exibicao": "Lima endodôntica 25 mm",
      "tipo_produto_codigo": "DESCARTAVEL_VARIACAO",
      "tipo_produto": "Descartável com variação",
      "grupo": "descartavel",
      "unidade_medida": "blister",
      "quantidade_atual": 6,
      "estoque_minimo": 4,
      "validade": null
    }
  ]
}
```

Erros: `400` id inválido · `404` categoria não encontrada.

#### `PUT /api/categories/{id}` — Renomear categoria

Mesmo corpo do cadastro.

Requisição:

```json
{ "nome": "Endodontia e Prótese" }
```

Resposta `200`:

```json
{
  "id": 12,
  "nome": "Endodontia e Prótese",
  "criado_em": "2026-09-20T14:02:10.118034-03:00",
  "atualizado_em": "2026-10-06T16:48:03.209554-03:00"
}
```

Erros: `404` categoria não encontrada · `409` nome já existe · `422` nome inválido.

#### `DELETE /api/categories/{id}` — Excluir categoria

Resposta `204` sem corpo.

Exclusão lógica (`excluido_em`): a categoria some das listagens e o nome pode ser reutilizado. Só é permitida se a categoria não tiver produtos ativos; produtos já excluídos não impedem.

Erros: `404` categoria não encontrada (ou já excluída) · `409` categoria com produtos ativos:

```json
{ "error": "a categoria possui produtos vinculados e não pode ser excluída" }
```

---

### Produtos

O saldo (`quantidade_atual`) nunca é gravado direto no produto: toda mudança gera um registro em `movimentacoes_estoque`, e o trigger do banco atualiza o saldo.

- No cadastro, `quantidade` vira uma movimentação de **entrada** `0 → quantidade`.
- Na edição, se `quantidade` for diferente do saldo atual, vira uma movimentação `saldo → quantidade`: **entrada** se o saldo sobe, **saída** se desce. Sem `quantidade`, o saldo não muda.

Tipos de produto (`tipo_produto_codigo`):

| Código | Validade | Variação |
| :--- | :--- | :--- |
| `PERECIVEL` | Obrigatória | Não permitida |
| `PERECIVEL_VARIACAO` | Obrigatória | Permitida (cor) |
| `DESCARTAVEL_VARIACAO` | Não permitida | Permitida (tamanho) |
| `DESCARTAVEL_SIMPLES` | Não permitida | Não permitida |
| `DURAVEL` | Não permitida | Não permitida |

Situação do estoque (`situacao`): `esgotado` (saldo 0) · `critico` (saldo ≤ 25% do mínimo) · `baixo` (saldo < mínimo) · `ok`.

#### Objeto produto (resposta)

```json
{
  "id": 517,
  "nome": "Ionômero de vidro",
  "variacao": "A3",
  "nome_exibicao": "Ionômero de vidro A3",
  "tipo_produto_codigo": "PERECIVEL_VARIACAO",
  "tipo_produto": "Perecível com variação",
  "grupo": "perecivel",
  "categoria_id": 12,
  "categoria": "Endodontia",
  "unidade_medida_id": 9,
  "unidade_medida": "frasco",
  "quantidade_atual": 8,
  "estoque_minimo": 3,
  "situacao": "ok",
  "faltam": 0,
  "validade": "2027-08-15",
  "dias_para_vencer": 309,
  "criado_em": "2026-10-10T11:20:45.602118-03:00",
  "atualizado_em": "2026-10-10T11:20:45.602118-03:00"
}
```

| Campo | Descrição |
| :--- | :--- |
| `nome_exibicao` | `nome` + `variacao` |
| `faltam` | Quanto falta para atingir o estoque mínimo (nunca negativo) |
| `dias_para_vencer` | Dias até a validade; negativo = vencido; `null` sem validade |

#### `GET /api/products` — Listar produtos

Produtos ativos ordenados por nome e variação. Resposta `200`: lista de [objetos produto](#objeto-produto-resposta).

```json
[
  {
    "id": 517,
    "nome": "Ionômero de vidro",
    "variacao": "A3",
    "nome_exibicao": "Ionômero de vidro A3",
    "tipo_produto_codigo": "PERECIVEL_VARIACAO",
    "tipo_produto": "Perecível com variação",
    "grupo": "perecivel",
    "categoria_id": 12,
    "categoria": "Endodontia",
    "unidade_medida_id": 9,
    "unidade_medida": "frasco",
    "quantidade_atual": 8,
    "estoque_minimo": 3,
    "situacao": "ok",
    "faltam": 0,
    "validade": "2027-08-15",
    "dias_para_vencer": 309,
    "criado_em": "2026-10-10T11:20:45.602118-03:00",
    "atualizado_em": "2026-10-10T11:20:45.602118-03:00"
  },
  {
    "id": 522,
    "nome": "Sugador descartável",
    "variacao": null,
    "nome_exibicao": "Sugador descartável",
    "tipo_produto_codigo": "DESCARTAVEL_SIMPLES",
    "tipo_produto": "Descartável simples",
    "grupo": "descartavel",
    "categoria_id": 14,
    "categoria": "Atendimento clínico",
    "unidade_medida_id": 10,
    "unidade_medida": "saco",
    "quantidade_atual": 1,
    "estoque_minimo": 6,
    "situacao": "critico",
    "faltam": 5,
    "validade": null,
    "dias_para_vencer": null,
    "criado_em": "2026-10-11T15:03:37.118400-03:00",
    "atualizado_em": "2026-10-12T09:41:08.553291-03:00"
  }
]
```

#### `GET /api/products/{id}` — Buscar produto

Resposta `200`: um [objeto produto](#objeto-produto-resposta).

Erros: `400` id inválido · `404` produto não encontrado (ou excluído).

#### `POST /api/products` — Cadastrar produto

| Campo | Tipo | Obrigatório | Regra |
| :--- | :--- | :--- | :--- |
| `tipo_produto_codigo` | string | Sim | Um dos códigos da tabela de tipos |
| `categoria_id` | inteiro | Sim | Categoria existente |
| `unidade_medida_id` | inteiro | Sim | Unidade existente |
| `nome` | string | Sim | Até 120 caracteres |
| `variacao` | string | Não | Até 40 caracteres; só se o tipo permitir |
| `estoque_minimo` | inteiro | Sim | ≥ 0 |
| `validade` | string | Depende do tipo | `AAAA-MM-DD` |
| `quantidade` | inteiro | Não | Saldo inicial, ≥ 0. Padrão `0` |
| `usuario_id` | inteiro | Não | Responsável pela movimentação de entrada |

`nome` + `variacao` deve ser único entre os produtos ativos (sem diferenciar maiúsculas).

Requisição:

```json
{
  "tipo_produto_codigo": "PERECIVEL_VARIACAO",
  "categoria_id": 12,
  "unidade_medida_id": 9,
  "nome": "Ionômero de vidro",
  "variacao": "A3",
  "estoque_minimo": 3,
  "validade": "2027-08-15",
  "quantidade": 8,
  "usuario_id": 42
}
```

Resposta `201`: o [objeto produto](#objeto-produto-resposta) criado, com `quantidade_atual: 8`.

Erros:

- `409` nome + variação já existe
- `422` campo inválido ou inexistente:

```json
{
  "error": "dados inválidos",
  "fields": { "categoria_id": "não existe" }
}
```

- `422` regra do tipo de produto recusada pelo banco:

```json
{ "error": "Tipo \"Perecível\" exige data de validade." }
```

```json
{ "error": "Tipo \"Durável\" não admite variação." }
```

#### `PUT /api/products/{id}` — Editar produto

Mesmos campos do cadastro (substitui todos os dados cadastrais). `quantidade` é opcional: se informada e diferente do saldo atual, gera uma entrada ou uma saída.

Requisição (retira 6 unidades — saída 8 → 2 — e muda o mínimo para 5):

```json
{
  "tipo_produto_codigo": "PERECIVEL_VARIACAO",
  "categoria_id": 12,
  "unidade_medida_id": 9,
  "nome": "Ionômero de vidro",
  "variacao": "A3",
  "estoque_minimo": 5,
  "validade": "2027-08-15",
  "quantidade": 2,
  "usuario_id": 42
}
```

Resposta `200`:

```json
{
  "id": 517,
  "nome": "Ionômero de vidro",
  "variacao": "A3",
  "nome_exibicao": "Ionômero de vidro A3",
  "tipo_produto_codigo": "PERECIVEL_VARIACAO",
  "tipo_produto": "Perecível com variação",
  "grupo": "perecivel",
  "categoria_id": 12,
  "categoria": "Endodontia",
  "unidade_medida_id": 9,
  "unidade_medida": "frasco",
  "quantidade_atual": 2,
  "estoque_minimo": 5,
  "situacao": "baixo",
  "faltam": 3,
  "validade": "2027-08-15",
  "dias_para_vencer": 307,
  "criado_em": "2026-10-10T11:20:45.602118-03:00",
  "atualizado_em": "2026-10-12T16:05:19.004873-03:00"
}
```

Erros: os mesmos do cadastro, mais `404` produto não encontrado (ou excluído).

#### `DELETE /api/products/{id}` — Excluir produto

Exclusão lógica: preenche `excluido_em` e o produto some das listagens, mas o histórico de movimentações é preservado. Depois de excluído, outro produto pode usar o mesmo nome e variação.

Resposta `204` sem corpo.

Erros: `400` id inválido · `404` produto não encontrado (ou já excluído).

---

### Idiomas

#### `GET /api/languages` — Listar idiomas

Todos os idiomas cadastrados, ordenados por nome.

Resposta `200`:

```json
[
  {
    "codigo": "fr-FR",
    "nome": "Français",
    "regiao": "France",
    "ativo": true,
    "criado_em": "2026-09-01T08:00:00-03:00"
  },
  {
    "codigo": "it-IT",
    "nome": "Italiano",
    "regiao": "Italia",
    "ativo": false,
    "criado_em": "2026-09-01T08:00:00-03:00"
  }
]
```

---

### Unidades de medida

#### `GET /api/units` — Listar unidades de medida

Todas as unidades, ordenadas por nome (no singular; a pluralização fica no front).

Resposta `200`:

```json
[
  {
    "id": 11,
    "nome": "blister",
    "criado_em": "2026-09-01T08:00:00-03:00"
  },
  {
    "id": 9,
    "nome": "frasco",
    "criado_em": "2026-09-01T08:00:00-03:00"
  }
]
```

---

### Histórico de movimentações

#### `GET /api/movements` — Listar movimentações

Histórico da tela Histórico, da mais recente para a mais antiga. Junta `movimentacoes_estoque` com `produtos`, `unidades_medida` e `usuarios`.

| Parâmetro (query string) | Obrigatório | Descrição |
| :--- | :--- | :--- |
| `tipo` | Não | `entrada`, `saida` ou `ajuste`. Vazio ou `todas` = todos os tipos. A API só gera `entrada` e `saida`; `ajuste` existe no banco para registros antigos |
| `limite` | Não | Quantidade de registros, de 1 a 500. Padrão `100` |

Exemplos: `/api/movements` · `/api/movements?tipo=saida` · `/api/movements?tipo=entrada&limite=20`

Resposta `200`:

```json
[
  {
    "id": 88,
    "tipo": "saida",
    "produto_id": 517,
    "produto": "Ionômero de vidro A3",
    "produto_excluido": false,
    "quantidade_anterior": 8,
    "quantidade_posterior": 7,
    "delta": -1,
    "unidade_medida": "frasco",
    "usuario_id": 42,
    "usuario": "Marina Costa",
    "criado_em": "2026-10-14T09:40:12.551302-03:00"
  },
  {
    "id": 81,
    "tipo": "entrada",
    "produto_id": 522,
    "produto": "Sugador descartável",
    "produto_excluido": true,
    "quantidade_anterior": 0,
    "quantidade_posterior": 3,
    "delta": 3,
    "unidade_medida": "saco",
    "usuario_id": null,
    "usuario": null,
    "criado_em": "2026-10-11T15:03:37.118400-03:00"
  }
]
```

| Campo | Descrição |
| :--- | :--- |
| `produto` | Nome de exibição (`nome` + `variacao`) |
| `produto_excluido` | `true` se o produto foi excluído depois; a movimentação continua no histórico |
| `delta` | `quantidade_posterior - quantidade_anterior` (negativo = retirada) |
| `unidade_medida` | Unidade no momento da movimentação (não muda se o produto trocar de unidade depois) |
| `usuario_id` / `usuario` | Quem fez a alteração; `null` se não informado ou se o usuário foi removido. Para exibir "Você", compare `usuario_id` com o `id` do login |

Erros: `422` `tipo` ou `limite` inválidos:

```json
{
  "error": "dados inválidos",
  "fields": { "tipo": "use entrada, saida ou ajuste" }
}
```
