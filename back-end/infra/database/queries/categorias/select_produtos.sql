-- Produtos ativos de uma categoria, com tipo e unidade de medida.
-- $1 = categoria_id
SELECT  p.id,
        p.nome,
        p.variacao,
        btrim(p.nome || ' ' || coalesce(p.variacao, '')) AS nome_exibicao,
        p.tipo_produto_codigo,
        t.nome                                           AS tipo_produto,
        t.grupo,
        um.nome                                          AS unidade_medida,
        p.quantidade_atual,
        p.estoque_minimo,
        to_char(p.validade, 'YYYY-MM-DD')                AS validade
FROM produtos p
JOIN tipos_produto   t  ON t.codigo = p.tipo_produto_codigo
JOIN unidades_medida um ON um.id    = p.unidade_medida_id
WHERE p.categoria_id = $1
  AND p.excluido_em IS NULL
ORDER BY p.nome, p.variacao NULLS FIRST;
