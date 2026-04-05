-- Migration: Create user_preferences table for smart preference learning
-- Stores per-source user preference profiles learned from chat conversations

CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    source_id INT REFERENCES sources(id) ON DELETE CASCADE,
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(source_id)
);

-- Also support a global preference (source_id = NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_preferences_global ON user_preferences (source_id) WHERE source_id IS NULL;
