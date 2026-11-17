UPDATE categorias c
SET excluido_em = now()
WHERE c.id = $1
  AND c.excluido_em IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM produtos p
      WHERE p.categoria_id = c.id AND p.excluido_em IS NULL
  )
RETURNING c.id;
