-- Lista as categorias com a quantidade de produtos ativos em cada uma.
SELECT  c.id,
        c.nome,
        count(p.id) AS itens,
        c.criado_em,
        c.atualizado_em
FROM categorias c
LEFT JOIN produtos p ON p.categoria_id = c.id AND p.excluido_em IS NULL
WHERE c.excluido_em IS NULL
GROUP BY c.id
ORDER BY c.nome;
