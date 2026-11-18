-- Atualiza nome, e-mail e idioma. atualizado_em é preenchido pelo trigger trg_usuarios_atualizado_em.
-- $1 = id, $2 = nome, $3 = email, $4 = idioma_codigo
UPDATE usuarios
SET nome          = $2,
    email         = $3,
    idioma_codigo = $4
WHERE id = $1
RETURNING id, nome, email, idioma_codigo, criado_em, atualizado_em;
