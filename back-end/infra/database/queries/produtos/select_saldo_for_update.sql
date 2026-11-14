-- Trava a linha do produto e lê o saldo atual (usado antes de editar).
-- $1 = id
SELECT quantidade_atual
FROM produtos
WHERE id = $1
  AND excluido_em IS NULL
FOR UPDATE;
