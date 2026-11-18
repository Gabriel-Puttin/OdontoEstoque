-- Histórico de movimentações, da mais recente para a mais antiga.
--   produtos:        inclui produtos excluídos (logicamente),
--                    pois o histórico é permanente.
--   unidades_medida: usa o snapshot gravado na movimentação, não a unidade atual do produto.
--   usuarios:        LEFT JOIN, pois usuario_id é opcional (NULL se o usuário foi removido).
-- $1 = tipo ('entrada' | 'saida' | 'ajuste'; NULL = todos), $2 = limite de linhas
SELECT  m.id,
        m.tipo,
        m.produto_id,
        btrim(p.nome || ' ' || coalesce(p.variacao, ''))  AS produto,
        (p.excluido_em IS NOT NULL)                       AS produto_excluido,
        m.quantidade_anterior,
        m.quantidade_posterior,
        m.delta,
        um.nome                                           AS unidade_medida,
        m.usuario_id,
        u.nome                                            AS usuario,
        m.criado_em
FROM movimentacoes_estoque m
JOIN produtos        p  ON p.id  = m.produto_id
JOIN unidades_medida um ON um.id = m.unidade_medida_id
LEFT JOIN usuarios   u  ON u.id  = m.usuario_id
WHERE ($1::varchar IS NULL OR m.tipo = $1)
ORDER BY m.criado_em DESC, m.id DESC
LIMIT $2;
