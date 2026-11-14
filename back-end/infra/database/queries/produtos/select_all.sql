-- Lista os produtos ativos com tipo, categoria, unidade e situação do estoque.
-- Situação: esgotado (0) | critico (<= 25% do mínimo) | baixo (< mínimo) | ok.
SELECT  p.id,
        p.nome,
        p.variacao,
        btrim(p.nome || ' ' || coalesce(p.variacao, ''))  AS nome_exibicao,
        p.tipo_produto_codigo,
        t.nome                                            AS tipo_produto,
        t.grupo,
        p.categoria_id,
        c.nome                                            AS categoria,
        p.unidade_medida_id,
        um.nome                                           AS unidade_medida,
        p.quantidade_atual,
        p.estoque_minimo,
        CASE
            WHEN p.quantidade_atual = 0                     THEN 'esgotado'
            WHEN p.quantidade_atual * 4 <= p.estoque_minimo THEN 'critico'
            WHEN p.quantidade_atual < p.estoque_minimo      THEN 'baixo'
            ELSE 'ok'
        END                                               AS situacao,
        GREATEST(p.estoque_minimo - p.quantidade_atual, 0) AS faltam,
        to_char(p.validade, 'YYYY-MM-DD')                 AS validade,
        (p.validade - CURRENT_DATE)                       AS dias_para_vencer,
        p.criado_em,
        p.atualizado_em
FROM produtos p
JOIN tipos_produto   t  ON t.codigo = p.tipo_produto_codigo
JOIN categorias      c  ON c.id     = p.categoria_id
JOIN unidades_medida um ON um.id    = p.unidade_medida_id
WHERE p.excluido_em IS NULL
ORDER BY p.nome, p.variacao NULLS FIRST;
