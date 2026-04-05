-- Migration: Create notification_settings table for user-configurable notification rules
-- Single-row table storing global notification preferences

CREATE TABLE IF NOT EXISTS notification_settings (
    id SERIAL PRIMARY KEY,
    min_innovation_score INT NOT NULL DEFAULT 8,
    min_depth_score INT NOT NULL DEFAULT 7,
    notify_on_interesting BOOLEAN NOT NULL DEFAULT true,
    watched_source_ids JSONB DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default settings row if not exists
INSERT INTO notification_settings (id, min_innovation_score, min_depth_score, notify_on_interesting, watched_source_ids, enabled)
VALUES (1, 8, 7, true, '[]', true)
ON CONFLICT (id) DO NOTHING;
