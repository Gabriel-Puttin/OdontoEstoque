-- =====================================================================
-- OdontoEstoque - schema.sql
-- Compatível com PostgreSQL 13+ (executável em banco vazio)
-- Execução: psql -v ON_ERROR_STOP=1 -d odontoestoque -f schema.sql
-- =====================================================================
BEGIN;

-- Busca por nome (ILIKE '%texto%') na tela Estoque
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------------
-- Função utilitária: mantém atualizado_em
-- ---------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_set_atualizado_em()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.atualizado_em := now();
    RETURN NEW;
END;
$$;

-- =====================================================================
-- 1. TABELAS DE DOMÍNIO
-- =====================================================================

-- Idiomas disponíveis (tela Idioma)
CREATE TABLE idiomas (
    codigo      varchar(10)  PRIMARY KEY
                CHECK (codigo ~ '^[a-z]{2}-[A-Z]{2}$'),
    nome        varchar(40)  NOT NULL,          -- Português, English, Español
    regiao      varchar(40)  NOT NULL,          -- Brasil, United States, España
    ativo       boolean      NOT NULL DEFAULT true,
    criado_em   timestamptz  NOT NULL DEFAULT now()
);
COMMENT ON TABLE idiomas IS 'Idiomas oferecidos pelo aplicativo (tela Idioma). PK natural: código BCP-47 (pt-BR).';

-- Categorias (tela Categorias / formulário de item)
CREATE TABLE categorias (
    id             smallint     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nome           varchar(80)  NOT NULL CHECK (btrim(nome) <> ''),
    criado_em      timestamptz  NOT NULL DEFAULT now(),
    atualizado_em  timestamptz  NOT NULL DEFAULT now(),
    excluido_em    timestamptz                      -- exclusão lógica; NULL = ativa
);
-- Único só entre categorias ativas: o nome de uma categoria excluída pode ser reutilizado
CREATE UNIQUE INDEX uq_categorias_nome ON categorias (lower(nome)) WHERE excluido_em IS NULL;
COMMENT ON TABLE categorias IS 'Categorias de produtos. A contagem de itens por categoria é calculada (vw_categorias_itens).';
COMMENT ON COLUMN categorias.excluido_em IS 'Preenchido ao excluir a categoria (só permitido sem produtos ativos); NULL = ativa.';
CREATE TRIGGER trg_categorias_atualizado_em BEFORE UPDATE ON categorias
    FOR EACH ROW EXECUTE FUNCTION fn_set_atualizado_em();

-- Unidades de medida (caixa, unidade, tubo, pacote)
CREATE TABLE unidades_medida (
    id         smallint     GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nome       varchar(30)  NOT NULL CHECK (btrim(nome) <> ''),
    criado_em  timestamptz  NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_unidades_medida_nome ON unidades_medida (lower(nome));
COMMENT ON TABLE unidades_medida IS 'Unidades de medida em singular (a pluralização é responsabilidade da aplicação/i18n).';

-- Tipos de produto: definem as regras de validade e variação
CREATE TABLE tipos_produto (
    codigo          varchar(30)  PRIMARY KEY CHECK (codigo ~ '^[A-Z_]+$'),
    nome            varchar(60)  NOT NULL UNIQUE,
    descricao       varchar(200) NOT NULL,
    grupo           varchar(15)  NOT NULL CHECK (grupo IN ('perecivel','descartavel','duravel')),
    exige_validade  boolean      NOT NULL,
    tipo_variacao   varchar(10)  CHECK (tipo_variacao IN ('cor','tamanho')), -- NULL = sem variação
    ordem           smallint     NOT NULL DEFAULT 0,
    -- perecível <=> exige validade
    CONSTRAINT ck_tipo_grupo_validade CHECK ((grupo = 'perecivel') = exige_validade)
);
COMMENT ON TABLE tipos_produto IS 'Tipos de produto e suas regras. "grupo" alimenta os filtros Perecível/Descartável/Durável da tela Estoque.';
COMMENT ON COLUMN tipos_produto.tipo_variacao IS 'Natureza da variação (cor ou tamanho); NULL quando o tipo não admite variação.';

-- =====================================================================
-- 2. USUÁRIOS
-- =====================================================================
CREATE TABLE usuarios (
    id             bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nome           varchar(120) NOT NULL CHECK (btrim(nome) <> ''),
    email          varchar(254) NOT NULL CHECK (email ~* '^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$'),
    senha_hash     text         NOT NULL CHECK (char_length(senha_hash) >= 50), -- bcrypt/argon2; nunca texto puro
    idioma_codigo  varchar(10)  NOT NULL DEFAULT 'pt-BR'
                   REFERENCES idiomas (codigo) ON UPDATE CASCADE ON DELETE RESTRICT,
    criado_em      timestamptz  NOT NULL DEFAULT now(),
    atualizado_em  timestamptz  NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_usuarios_email ON usuarios (lower(email));
COMMENT ON TABLE usuarios IS 'Usuários (cadastro, login, perfil). Confirmação de senha e inicial do avatar não são persistidas.';
COMMENT ON COLUMN usuarios.senha_hash IS 'Hash da senha gerado pela aplicação (bcrypt/argon2id). Mínimo de 8 caracteres é validado na aplicação, antes do hash.';
CREATE TRIGGER trg_usuarios_atualizado_em BEFORE UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION fn_set_atualizado_em();

-- =====================================================================
-- 3. PRODUTOS
-- =====================================================================
CREATE TABLE produtos (
    id                  bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tipo_produto_codigo varchar(30)  NOT NULL REFERENCES tipos_produto (codigo) ON UPDATE CASCADE ON DELETE RESTRICT,
    categoria_id        smallint     NOT NULL REFERENCES categorias (id)        ON UPDATE CASCADE ON DELETE RESTRICT,
    unidade_medida_id   smallint     NOT NULL REFERENCES unidades_medida (id)   ON UPDATE CASCADE ON DELETE RESTRICT,
    nome                varchar(120) NOT NULL CHECK (btrim(nome) <> ''),
    variacao            varchar(40)  CHECK (variacao IS NULL OR btrim(variacao) <> ''), -- cor (A2) ou tamanho (M)
    quantidade_atual    integer      NOT NULL DEFAULT 0 CHECK (quantidade_atual >= 0),
    estoque_minimo      integer      NOT NULL CHECK (estoque_minimo >= 0),
    validade            date,
    criado_em           timestamptz  NOT NULL DEFAULT now(),
    atualizado_em       timestamptz  NOT NULL DEFAULT now(),
    excluido_em         timestamptz                      -- exclusão lógica (preserva o histórico)
);
COMMENT ON TABLE produtos IS 'Itens de estoque. Nome exibido = nome + variacao. Exclusão é lógica para preservar movimentações.';
COMMENT ON COLUMN produtos.quantidade_atual IS 'Saldo corrente, mantido EXCLUSIVAMENTE por movimentacoes_estoque (trigger). Cache justificado: leitura frequente em todas as telas.';
COMMENT ON COLUMN produtos.validade IS 'Obrigatória para tipos com exige_validade; proibida nos demais (trigger).';
COMMENT ON COLUMN produtos.excluido_em IS 'Preenchido ao "excluir" o item na interface; NULL = ativo.';

CREATE UNIQUE INDEX uq_produtos_nome_variacao ON produtos (lower(nome), lower(coalesce(variacao, '')))
    WHERE excluido_em IS NULL;
CREATE INDEX ix_produtos_categoria ON produtos (categoria_id) WHERE excluido_em IS NULL;
CREATE INDEX ix_produtos_tipo      ON produtos (tipo_produto_codigo) WHERE excluido_em IS NULL;
CREATE INDEX ix_produtos_validade  ON produtos (validade) WHERE excluido_em IS NULL AND validade IS NOT NULL;
CREATE INDEX ix_produtos_nome_trgm ON produtos USING gin (nome gin_trgm_ops);

-- Regras que dependem do tipo do produto + proteção do saldo
CREATE OR REPLACE FUNCTION fn_produtos_regras()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_tipo tipos_produto%ROWTYPE;
BEGIN
    SELECT * INTO v_tipo FROM tipos_produto WHERE codigo = NEW.tipo_produto_codigo;

    IF v_tipo.exige_validade AND NEW.validade IS NULL THEN
        RAISE EXCEPTION 'Tipo "%" exige data de validade.', v_tipo.nome USING ERRCODE = 'check_violation';
    END IF;
    IF v_tipo.exige_validade IS FALSE AND NEW.validade IS NOT NULL THEN
        RAISE EXCEPTION 'Tipo "%" não possui validade.', v_tipo.nome USING ERRCODE = 'check_violation';
    END IF;
    IF v_tipo.codigo IS NOT NULL AND v_tipo.tipo_variacao IS NULL AND NEW.variacao IS NOT NULL THEN
        RAISE EXCEPTION 'Tipo "%" não admite variação.', v_tipo.nome USING ERRCODE = 'check_violation';
    END IF;

    IF TG_OP = 'INSERT' THEN
        IF NEW.quantidade_atual <> 0 THEN
            RAISE EXCEPTION 'Crie o produto com saldo 0 e registre uma movimentação de entrada.' USING ERRCODE = 'check_violation';
        END IF;
    ELSE
        -- saldo só pode mudar via trigger de movimentação (profundidade >= 2)
        IF NEW.quantidade_atual IS DISTINCT FROM OLD.quantidade_atual AND pg_trigger_depth() < 2 THEN
            RAISE EXCEPTION 'quantidade_atual só pode ser alterada por movimentacoes_estoque.' USING ERRCODE = 'check_violation';
        END IF;
        NEW.atualizado_em := now();
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_produtos_regras BEFORE INSERT OR UPDATE ON produtos
    FOR EACH ROW EXECUTE FUNCTION fn_produtos_regras();

-- =====================================================================
-- 4. MOVIMENTAÇÕES DE ESTOQUE (histórico permanente)
-- =====================================================================
CREATE TABLE movimentacoes_estoque (
    id                  bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    produto_id          bigint       NOT NULL REFERENCES produtos (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    usuario_id          bigint       REFERENCES usuarios (id) ON UPDATE CASCADE ON DELETE SET NULL,
    tipo                varchar(10)  NOT NULL CHECK (tipo IN ('entrada','saida','ajuste')),
    quantidade_anterior integer      NOT NULL CHECK (quantidade_anterior >= 0),
    quantidade_posterior integer     NOT NULL CHECK (quantidade_posterior >= 0),
    delta               integer      GENERATED ALWAYS AS (quantidade_posterior - quantidade_anterior) STORED,
    unidade_medida_id   smallint     NOT NULL REFERENCES unidades_medida (id) ON UPDATE CASCADE ON DELETE RESTRICT,
    criado_em           timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT ck_mov_tipo_sinal CHECK (
        (tipo = 'entrada' AND delta > 0) OR
        (tipo = 'saida'   AND delta < 0) OR
        (tipo = 'ajuste'  AND delta <> 0)
    )
);
COMMENT ON TABLE movimentacoes_estoque IS 'Registro imutável de toda alteração de quantidade (tela Histórico). Ex.: "Luvas M  -1 caixa (12 -> 11)".';
COMMENT ON COLUMN movimentacoes_estoque.delta IS 'Quantidade movimentada com sinal (gerada): posterior - anterior.';
COMMENT ON COLUMN movimentacoes_estoque.unidade_medida_id IS 'Snapshot da unidade no momento da movimentação (preenchido por trigger).';
COMMENT ON COLUMN movimentacoes_estoque.usuario_id IS 'Responsável; NULL se o usuário for removido ("quando aplicável").';

CREATE INDEX ix_mov_produto_data ON movimentacoes_estoque (produto_id, criado_em DESC);
CREATE INDEX ix_mov_data         ON movimentacoes_estoque (criado_em DESC);
CREATE INDEX ix_mov_tipo_data    ON movimentacoes_estoque (tipo, criado_em DESC);
CREATE INDEX ix_mov_usuario      ON movimentacoes_estoque (usuario_id) WHERE usuario_id IS NOT NULL;

-- Valida o saldo e aplica a movimentação ao produto (atomicamente)
CREATE OR REPLACE FUNCTION fn_mov_aplicar()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    v_prod produtos%ROWTYPE;
BEGIN
    SELECT * INTO v_prod FROM produtos WHERE id = NEW.produto_id FOR UPDATE;
    IF NOT FOUND OR v_prod.excluido_em IS NOT NULL THEN
        RAISE EXCEPTION 'Produto % inexistente ou excluído.', NEW.produto_id USING ERRCODE = 'foreign_key_violation';
    END IF;
    IF NEW.quantidade_anterior <> v_prod.quantidade_atual THEN
        RAISE EXCEPTION 'Saldo desatualizado: informado %, atual %.', NEW.quantidade_anterior, v_prod.quantidade_atual
            USING ERRCODE = 'serialization_failure';
    END IF;
    NEW.unidade_medida_id := v_prod.unidade_medida_id;
    UPDATE produtos SET quantidade_atual = NEW.quantidade_posterior WHERE id = NEW.produto_id;
    RETURN NEW;
END;
$$;
CREATE TRIGGER trg_mov_aplicar BEFORE INSERT ON movimentacoes_estoque
    FOR EACH ROW EXECUTE FUNCTION fn_mov_aplicar();

-- Imutabilidade: só permite anular o responsável (ON DELETE SET NULL de usuarios)
CREATE OR REPLACE FUNCTION fn_mov_imutavel()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND NEW.usuario_id IS NULL
       AND ROW(NEW.id, NEW.produto_id, NEW.tipo, NEW.quantidade_anterior, NEW.quantidade_posterior, NEW.unidade_medida_id, NEW.criado_em)
           IS NOT DISTINCT FROM
           ROW(OLD.id, OLD.produto_id, OLD.tipo, OLD.quantidade_anterior, OLD.quantidade_posterior, OLD.unidade_medida_id, OLD.criado_em)
    THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'movimentacoes_estoque é um registro permanente (somente inserção).' USING ERRCODE = 'restrict_violation';
END;
$$;
CREATE TRIGGER trg_mov_imutavel BEFORE UPDATE OR DELETE ON movimentacoes_estoque
    FOR EACH ROW EXECUTE FUNCTION fn_mov_imutavel();

-- =====================================================================
-- 5. VIEWS (valores calculados - nada disso é armazenado)
-- =====================================================================

-- Situação do estoque e validade. Regra (hipótese a confirmar):
--   esgotado: qtd = 0 | crítico: qtd <= 25% do mínimo | baixo: qtd < mínimo | ok: qtd >= mínimo
CREATE VIEW vw_produtos_situacao AS
SELECT  p.id,
        p.nome,
        p.variacao,
        btrim(p.nome || ' ' || coalesce(p.variacao, ''))  AS nome_exibicao,
        p.tipo_produto_codigo,
        t.grupo,
        p.categoria_id,
        c.nome                                            AS categoria,
        um.nome                                           AS unidade_medida,
        p.quantidade_atual,
        p.estoque_minimo,
        CASE
            WHEN p.quantidade_atual = 0                      THEN 'esgotado'
            WHEN p.quantidade_atual * 4 <= p.estoque_minimo  THEN 'critico'
            WHEN p.quantidade_atual < p.estoque_minimo       THEN 'baixo'
            ELSE 'ok'
        END                                               AS situacao,
        GREATEST(p.estoque_minimo - p.quantidade_atual, 0) AS faltam,
        p.validade,
        (p.validade - CURRENT_DATE)                       AS dias_para_vencer  -- negativo = vencido
FROM produtos p
JOIN tipos_produto t  ON t.codigo = p.tipo_produto_codigo
JOIN categorias c     ON c.id = p.categoria_id
JOIN unidades_medida um ON um.id = p.unidade_medida_id
WHERE p.excluido_em IS NULL;
COMMENT ON VIEW vw_produtos_situacao IS 'Estoque, Reposição, Validades e Dashboard: situação, faltantes e dias para vencer.';

CREATE VIEW vw_categorias_itens AS
SELECT  c.id, c.nome, count(p.id) AS itens
FROM categorias c
LEFT JOIN produtos p ON p.categoria_id = c.id AND p.excluido_em IS NULL
WHERE c.excluido_em IS NULL
GROUP BY c.id, c.nome;
COMMENT ON VIEW vw_categorias_itens IS 'Tela Categorias: "N itens" por categoria.';

-- =====================================================================
-- 6. DADOS INICIAIS (presentes nos protótipos)
-- =====================================================================
INSERT INTO idiomas (codigo, nome, regiao) VALUES
    ('pt-BR', 'Português', 'Brasil'),
    ('en-US', 'English',   'United States'),
    ('es-ES', 'Español',   'España')
ON CONFLICT DO NOTHING;

INSERT INTO tipos_produto (codigo, nome, descricao, grupo, exige_validade, tipo_variacao, ordem) VALUES
    ('PERECIVEL',            'Perecível',            'Validade obrigatória; sem variação.',     'perecivel',   true,  NULL,      1),
    ('PERECIVEL_VARIACAO',   'Perecível com variação','Validade obrigatória; variação por cor.', 'perecivel',   true,  'cor',     2),
    ('DESCARTAVEL_VARIACAO', 'Descartável com variação','Sem validade; variação por tamanho.',   'descartavel', false, 'tamanho', 3),
    ('DESCARTAVEL_SIMPLES',  'Descartável simples',  'Sem validade e sem variação.',            'descartavel', false, NULL,      4),
    ('DURAVEL',              'Durável',              'Sem validade e sem variação.',            'duravel',     false, NULL,      5)
ON CONFLICT DO NOTHING;

INSERT INTO unidades_medida (nome) VALUES ('caixa'), ('unidade'), ('tubo'), ('pacote')
ON CONFLICT DO NOTHING;

INSERT INTO categorias (nome) VALUES
    ('Biossegurança e EPI'), ('Esterilização'), ('Descartáveis do atendimento'), ('Moldagem'),
    ('Anestésicos'), ('Materiais restauradores e preventivos'), ('Instrumentos rotatórios'),
    ('Higiene e itens para o paciente')
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------
-- Usuários de demonstração (senha_hash = bcrypt, custo 10)
--   Ana  -> ana@clinica.com  / 12345678
--   João -> joão@clinica.com / 123456789
-- ---------------------------------------------------------------------
INSERT INTO usuarios (nome, email, senha_hash) VALUES
    ('Ana',  'ana@clinica.com',  '$2a$10$I0jhBvfh55Xcma.wIUS9ru0UwuHaf/NsNTPSvMjbAa/orXjqIvkGe'),
    ('João', 'joão@clinica.com', '$2a$10$C/hm50F1uY81JNLtT7uwyu3.3nnRf8RgEXr28.XJw8e5nqKM1iXt6')
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------
-- Produtos de demonstração
-- O trigger fn_produtos_regras exige que o produto nasça com saldo 0. Por isso:
--   1. os produtos são inseridos com quantidade_atual = 0;
--   2. o saldo inicial entra como movimentação de 'entrada' (Ana, há 15 dias);
--   3. algumas 'saidas' (João/Ana) populam a tela Histórico.
-- validade_dias = dias a partir da execução do script (negativo = já vencido).
-- ---------------------------------------------------------------------
CREATE TEMP TABLE seed_produtos (
    categoria      varchar(80),
    tipo           varchar(30),
    unidade        varchar(30),
    nome           varchar(120),
    variacao       varchar(40),
    estoque_minimo integer,
    validade_dias  integer,
    quantidade     integer
) ON COMMIT DROP;

INSERT INTO seed_produtos VALUES
    -- Biossegurança e EPI
    ('Biossegurança e EPI', 'DESCARTAVEL_VARIACAO', 'caixa',   'Luvas de procedimento',              'P',            5,  NULL,  8),
    ('Biossegurança e EPI', 'DESCARTAVEL_VARIACAO', 'caixa',   'Luvas de procedimento',              'M',            8,  NULL, 12),
    ('Biossegurança e EPI', 'DESCARTAVEL_VARIACAO', 'caixa',   'Luvas de procedimento',              'G',            5,  NULL,  2),
    ('Biossegurança e EPI', 'DESCARTAVEL_SIMPLES',  'caixa',   'Máscara descartável tripla',         NULL,           4,  NULL, 10),
    ('Biossegurança e EPI', 'DESCARTAVEL_SIMPLES',  'pacote',  'Gorro descartável',                  NULL,           3,  NULL,  0),
    ('Biossegurança e EPI', 'DURAVEL',              'unidade', 'Óculos de proteção',                 NULL,           2,  NULL,  4),
    -- Esterilização
    ('Esterilização', 'DESCARTAVEL_VARIACAO', 'caixa',   'Envelope para autoclave',            '90 x 260 mm',  3,  NULL,  6),
    ('Esterilização', 'DESCARTAVEL_VARIACAO', 'caixa',   'Envelope para autoclave',            '200 x 400 mm', 2,  NULL,  1),
    ('Esterilização', 'PERECIVEL',            'caixa',   'Indicador biológico',                NULL,           1,   200,  2),
    ('Esterilização', 'PERECIVEL',            'unidade', 'Detergente enzimático 1 L',          NULL,           2,   365,  3),
    ('Esterilização', 'DESCARTAVEL_SIMPLES',  'unidade', 'Fita para autoclave',                NULL,           3,  NULL,  5),
    -- Descartáveis do atendimento
    ('Descartáveis do atendimento', 'DESCARTAVEL_SIMPLES', 'pacote', 'Gaze estéril',              NULL, 5, NULL, 2),
    ('Descartáveis do atendimento', 'DESCARTAVEL_SIMPLES', 'pacote', 'Sugador descartável',       NULL, 4, NULL, 9),
    ('Descartáveis do atendimento', 'DESCARTAVEL_SIMPLES', 'pacote', 'Babador descartável',       NULL, 3, NULL, 6),
    ('Descartáveis do atendimento', 'DESCARTAVEL_SIMPLES', 'pacote', 'Rolete de algodão',         NULL, 4, NULL, 1),
    ('Descartáveis do atendimento', 'DESCARTAVEL_SIMPLES', 'pacote', 'Copo descartável 180 ml',   NULL, 2, NULL, 5),
    -- Moldagem
    ('Moldagem', 'PERECIVEL', 'pacote',  'Alginato',                    NULL, 4,  150,  3),
    ('Moldagem', 'PERECIVEL', 'caixa',   'Silicone de condensação',     NULL, 1,  240,  2),
    ('Moldagem', 'PERECIVEL', 'pacote',  'Gesso pedra tipo IV',         NULL, 2,  300,  4),
    ('Moldagem', 'DURAVEL',   'unidade', 'Moldeira metálica perfurada', NULL, 6, NULL, 10),
    -- Anestésicos
    ('Anestésicos', 'PERECIVEL', 'caixa',   'Lidocaína 2% com epinefrina',        NULL, 3, 400, 6),
    ('Anestésicos', 'PERECIVEL', 'caixa',   'Mepivacaína 3% sem vasoconstritor',  NULL, 2,  25, 3),
    ('Anestésicos', 'PERECIVEL', 'caixa',   'Articaína 4% com epinefrina',        NULL, 2, 330, 1),
    ('Anestésicos', 'PERECIVEL', 'unidade', 'Anestésico tópico benzocaína 20%',   NULL, 2, -10, 2),
    -- Materiais restauradores e preventivos
    ('Materiais restauradores e preventivos', 'PERECIVEL_VARIACAO', 'tubo',    'Resina composta',    'A2', 2, 270, 3),
    ('Materiais restauradores e preventivos', 'PERECIVEL_VARIACAO', 'tubo',    'Resina composta',    'A3', 2, 270, 4),
    ('Materiais restauradores e preventivos', 'PERECIVEL_VARIACAO', 'tubo',    'Resina composta',    'B1', 1,  60, 1),
    ('Materiais restauradores e preventivos', 'PERECIVEL',          'unidade', 'Ionômero de vidro',  NULL, 2, 180, 3),
    ('Materiais restauradores e preventivos', 'PERECIVEL',          'unidade', 'Adesivo universal',  NULL, 1,  90, 2),
    ('Materiais restauradores e preventivos', 'PERECIVEL',          'unidade', 'Verniz fluoretado',  NULL, 2, 210, 0),
    -- Instrumentos rotatórios
    ('Instrumentos rotatórios', 'DURAVEL', 'unidade', 'Broca diamantada esférica 1012',       NULL, 5, NULL, 12),
    ('Instrumentos rotatórios', 'DURAVEL', 'unidade', 'Broca diamantada tronco-cônica 2135',  NULL, 5, NULL,  3),
    ('Instrumentos rotatórios', 'DURAVEL', 'unidade', 'Broca carbide 245',                    NULL, 4, NULL,  6),
    ('Instrumentos rotatórios', 'DURAVEL', 'unidade', 'Caneta de alta rotação',               NULL, 1, NULL,  2),
    ('Instrumentos rotatórios', 'DURAVEL', 'unidade', 'Contra-ângulo',                        NULL, 1, NULL,  1),
    -- Higiene e itens para o paciente
    ('Higiene e itens para o paciente', 'DESCARTAVEL_SIMPLES', 'unidade', 'Escova dental macia',                NULL, 20, NULL, 35),
    ('Higiene e itens para o paciente', 'DESCARTAVEL_SIMPLES', 'unidade', 'Fio dental 50 m',                    NULL, 15, NULL,  8),
    ('Higiene e itens para o paciente', 'PERECIVEL',           'tubo',    'Creme dental com flúor 90 g',        NULL, 10,  500, 14),
    ('Higiene e itens para o paciente', 'PERECIVEL',           'unidade', 'Enxaguante bucal sem álcool 500 ml', NULL,  6,  365,  4);

-- 1. Produtos (saldo 0)
INSERT INTO produtos (tipo_produto_codigo, categoria_id, unidade_medida_id, nome, variacao,
                      estoque_minimo, validade, criado_em, atualizado_em)
SELECT  s.tipo, c.id, um.id, s.nome, s.variacao,
        s.estoque_minimo, CURRENT_DATE + s.validade_dias,
        now() - interval '15 days', now() - interval '15 days'
FROM seed_produtos s
JOIN categorias      c  ON c.nome  = s.categoria
JOIN unidades_medida um ON um.nome = s.unidade;

-- 2. Saldo inicial: entrada 0 -> quantidade (o trigger fn_mov_aplicar atualiza produtos.quantidade_atual)
INSERT INTO movimentacoes_estoque (produto_id, usuario_id, tipo, quantidade_anterior, quantidade_posterior, criado_em)
SELECT  p.id, u.id, 'entrada', 0, s.quantidade, now() - interval '15 days'
FROM seed_produtos s
JOIN produtos p ON p.nome = s.nome AND p.variacao IS NOT DISTINCT FROM s.variacao
JOIN usuarios u ON u.email = 'ana@clinica.com'
WHERE s.quantidade > 0;

-- 3. Saídas posteriores (quantidade_anterior precisa bater com o saldo atual)
INSERT INTO movimentacoes_estoque (produto_id, usuario_id, tipo, quantidade_anterior, quantidade_posterior, criado_em)
SELECT  p.id, u.id, 'saida', s.anterior, s.posterior, now() - s.ha
FROM (VALUES
    ('Gaze estéril',                NULL, 2,  0,  'joão@clinica.com', interval '10 days'),
    ('Escova dental macia',         NULL, 35, 30, 'ana@clinica.com',  interval '6 days'),
    ('Lidocaína 2% com epinefrina', NULL, 6,  5,  'joão@clinica.com', interval '3 days'),
    ('Alginato',                    NULL, 3,  2,  'joão@clinica.com', interval '1 day'),
    ('Luvas de procedimento',       'M',  12, 11, 'joão@clinica.com', interval '3 hours')
) AS s (nome, variacao, anterior, posterior, email, ha)
JOIN produtos p ON p.nome = s.nome AND p.variacao IS NOT DISTINCT FROM s.variacao
JOIN usuarios u ON u.email = s.email;

COMMIT;