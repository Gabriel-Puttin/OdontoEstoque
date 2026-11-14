-- Registra uma movimentação. O trigger fn_mov_aplicar valida o saldo anterior,
-- preenche unidade_medida_id e atualiza produtos.quantidade_atual.
-- $1 = produto_id, $2 = usuario_id (opcional), $3 = tipo ('entrada' | 'saida' | 'ajuste'),
-- $4 = quantidade_anterior, $5 = quantidade_posterior
INSERT INTO movimentacoes_estoque (produto_id, usuario_id, tipo, quantidade_anterior, quantidade_posterior)
VALUES ($1, $2, $3, $4, $5);
