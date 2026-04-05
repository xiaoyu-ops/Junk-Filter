-- Migration: Create notifications table for high-value content alerts

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    content_id INT REFERENCES content(id) ON DELETE CASCADE,
    title VARCHAR(500),
    summary TEXT,
    innovation_score INT,
    depth_score INT,
    decision VARCHAR(50),
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications (is_read) WHERE is_read = false;
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications (created_at DESC);
