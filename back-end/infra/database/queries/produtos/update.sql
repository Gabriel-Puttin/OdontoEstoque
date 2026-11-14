-- Edita os dados cadastrais. quantidade_atual NÃO entra aqui (só muda por movimentação);
-- atualizado_em é preenchido pelo trigger fn_produtos_regras.
-- $1 = id, $2 = tipo_produto_codigo, $3 = categoria_id, $4 = unidade_medida_id,
-- $5 = nome, $6 = variacao, $7 = estoque_minimo, $8 = validade
UPDATE produtos
SET tipo_produto_codigo = $2,
    categoria_id        = $3,
    unidade_medida_id   = $4,
    nome                = $5,
    variacao            = $6,
    estoque_minimo      = $7,
    validade            = $8
WHERE id = $1
  AND excluido_em IS NULL;
