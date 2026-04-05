-- Migration: Add push_channels column to notification_settings
-- Stores configured push notification channels as JSONB array
-- Example: [{"type": "pushplus", "token": "xxx", "enabled": true}]

ALTER TABLE notification_settings ADD COLUMN IF NOT EXISTS push_channels JSONB DEFAULT '[]';
