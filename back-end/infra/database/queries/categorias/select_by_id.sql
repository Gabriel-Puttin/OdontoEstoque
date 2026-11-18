-- Busca uma categoria ativa pelo id.
-- $1 = id
SELECT id, nome, criado_em, atualizado_em
FROM categorias
WHERE id = $1
  AND excluido_em IS NULL;
