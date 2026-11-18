-- Exclusão lógica de uma categoria, permitida só se não houver produtos ativos nela.
-- Devolve 1 linha com tem_produtos:
--   nenhuma linha       -> categoria não existe (ou já excluída)
--   tem_produtos = true -> nada foi alterado; a categoria tem produtos ativos
--   tem_produtos = false-> categoria excluída
-- $1 = id
WITH alvo AS (
    SELECT  c.id,
            EXISTS (
                SELECT 1 FROM produtos p
                WHERE p.categoria_id = c.id AND p.excluido_em IS NULL
            ) AS tem_produtos
    FROM categorias c
    WHERE c.id = $1
      AND c.excluido_em IS NULL
), excluida AS (
    UPDATE categorias c
    SET excluido_em = now()
    FROM alvo
    WHERE c.id = alvo.id
      AND NOT alvo.tem_produtos
    RETURNING c.id
)
SELECT tem_produtos FROM alvo;
