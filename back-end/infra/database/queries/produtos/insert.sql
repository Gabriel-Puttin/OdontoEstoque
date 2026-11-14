-- Cria um produto. quantidade_atual nasce em 0 (exigido pelo trigger fn_produtos_regras);
-- o saldo inicial entra depois como movimentação de 'entrada'.
-- $1 = tipo_produto_codigo, $2 = categoria_id, $3 = unidade_medida_id, $4 = nome,
-- $5 = variacao, $6 = estoque_minimo, $7 = validade
INSERT INTO produtos (tipo_produto_codigo, categoria_id, unidade_medida_id, nome, variacao, estoque_minimo, validade)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;
