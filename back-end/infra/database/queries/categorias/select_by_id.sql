-- Busca uma categoria pelo id.
-- $1 = id
SELECT id, nome, criado_em, atualizado_em
FROM categorias
WHERE id = $1;
