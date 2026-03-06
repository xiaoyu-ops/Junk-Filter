-- Migration: Add threads table for sub-conversations
-- threads 表用于任务下的子对话管理

CREATE TABLE IF NOT EXISTS threads (
  id BIGSERIAL PRIMARY KEY,
  task_id BIGINT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
  title VARCHAR(200) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_threads_task_id ON threads(task_id);
CREATE INDEX IF NOT EXISTS idx_threads_updated_at ON threads(updated_at DESC);

-- Add thread_id column to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS thread_id BIGINT REFERENCES threads(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_messages_thread_id ON messages(thread_id);

-- Grant permissions
GRANT ALL PRIVILEGES ON threads TO junkfilter;
GRANT ALL PRIVILEGES ON threads_id_seq TO junkfilter;
