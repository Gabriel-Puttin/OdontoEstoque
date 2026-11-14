-- Exclusão lógica: preserva o produto e seu histórico de movimentações.
-- $1 = id
UPDATE produtos
SET excluido_em = now()
WHERE id = $1
  AND excluido_em IS NULL;
