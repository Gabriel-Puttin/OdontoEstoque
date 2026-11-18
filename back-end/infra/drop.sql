-- =====================================================================
-- OdontoEstoque - drop.sql
-- Remove TODOS os objetos criados por database.sql (tabelas, dados, views,
-- triggers e funções). Use para recriar o banco do zero:
--   1. rode este arquivo
--   2. rode database.sql
-- ATENÇÃO: apaga todos os dados de forma irreversível.
-- =====================================================================
BEGIN;

-- Views
DROP VIEW IF EXISTS vw_categorias_itens;
DROP VIEW IF EXISTS vw_produtos_situacao;

-- Tabelas (da que depende para a que é referenciada). Os triggers e
-- índices de cada tabela são removidos junto com ela.
DROP TABLE IF EXISTS movimentacoes_estoque CASCADE;
DROP TABLE IF EXISTS produtos              CASCADE;
DROP TABLE IF EXISTS usuarios              CASCADE;
DROP TABLE IF EXISTS tipos_produto         CASCADE;
DROP TABLE IF EXISTS unidades_medida       CASCADE;
DROP TABLE IF EXISTS categorias            CASCADE;
DROP TABLE IF EXISTS idiomas               CASCADE;

-- Funções dos triggers
DROP FUNCTION IF EXISTS fn_mov_imutavel();
DROP FUNCTION IF EXISTS fn_mov_aplicar();
DROP FUNCTION IF EXISTS fn_produtos_regras();
DROP FUNCTION IF EXISTS fn_set_atualizado_em();

-- A extensão pg_trgm é mantida (database.sql usa CREATE EXTENSION IF NOT EXISTS).

COMMIT;
