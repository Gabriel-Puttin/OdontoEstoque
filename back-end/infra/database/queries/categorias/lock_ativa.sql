-- Confirma que a categoria existe e está ativa, travando-a (FOR SHARE) até o fim da
-- transação: impede que ela seja excluída enquanto um produto é gravado nela.
-- $1 = id
SELECT id
FROM categorias
WHERE id = $1
  AND excluido_em IS NULL
FOR SHARE;
