-- Cria um usuário. O idioma usa o DEFAULT da tabela ('pt-BR') quando $4 for NULL.
-- $1 = nome, $2 = email, $3 = senha_hash, $4 = idioma_codigo (opcional)
INSERT INTO usuarios (nome, email, senha_hash, idioma_codigo)
VALUES ($1, $2, $3, COALESCE($4, 'pt-BR'))
RETURNING id, nome, email, idioma_codigo, criado_em, atualizado_em;
