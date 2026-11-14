-- Busca um usuário pelo e-mail, sem diferenciar maiúsculas (usa o índice uq_usuarios_email).
-- $1 = email
SELECT id, nome, email, senha_hash, idioma_codigo, criado_em, atualizado_em
FROM usuarios
WHERE lower(email) = lower($1);
