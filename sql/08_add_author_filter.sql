-- Migration: Add author_filter column to sources table
-- Stores author filtering rules as JSONB: {"mode": "whitelist|blacklist", "authors": ["name1", "name2"]}

ALTER TABLE sources ADD COLUMN IF NOT EXISTS author_filter JSONB DEFAULT '{}';
