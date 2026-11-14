-- Remove uma categoria. Falha com foreign_key_violation (23503) se algum produto a referenciar.
-- $1 = id
DELETE FROM categorias
WHERE id = $1;
