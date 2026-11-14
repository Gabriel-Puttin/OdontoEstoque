-- Renomeia uma categoria. atualizado_em é preenchido pelo trigger trg_categorias_atualizado_em.
-- $1 = id, $2 = nome
UPDATE categorias
SET nome = $2
WHERE id = $1
RETURNING id, nome, criado_em, atualizado_em;
