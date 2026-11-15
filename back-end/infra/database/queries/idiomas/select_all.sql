-- Lista todos os idiomas cadastrados (ativos e inativos).
SELECT codigo, nome, regiao, ativo, criado_em
FROM idiomas
ORDER BY nome;
