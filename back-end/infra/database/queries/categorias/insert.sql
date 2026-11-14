-- Cria uma categoria.
-- $1 = nome
INSERT INTO categorias (nome)
VALUES ($1)
RETURNING id, nome, criado_em, atualizado_em;
