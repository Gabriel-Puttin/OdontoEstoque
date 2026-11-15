-- Lista todas as unidades de medida (nomes no singular; a pluralização fica no front/i18n).
SELECT id, nome, criado_em
FROM unidades_medida
ORDER BY nome;
