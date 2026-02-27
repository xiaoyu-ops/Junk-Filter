-- Migration: Enhance messages table for Task-specific Chat (Agent Tuning)
-- This migration converts the generic messages table into a structured chat system
-- that supports Agent state snapshots and evaluation card references

-- Step 1: Add new columns to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_type VARCHAR(50) DEFAULT 'text';
-- message_type: 'user_query', 'ai_reply', 'system_notification'

ALTER TABLE messages ADD COLUMN IF NOT EXISTS agent_config JSONB;
-- 快照：对话时的 Agent 配置状态（temperature, topP, filter_rules 等）

ALTER TABLE messages ADD COLUMN IF NOT EXISTS referenced_card_ids BIGINT[];
-- 引用的评估卡片 ID（如用户问"为什么第 3 张卡片是 SKIP"）

ALTER TABLE messages ADD COLUMN IF NOT EXISTS parameter_updates JSONB;
-- 如果用户在聊天中修改了 Agent 参数，记录在这里

ALTER TABLE messages ADD COLUMN IF NOT EXISTS context_snapshot JSONB;
-- 完整的上下文快照（用于查询时重建对话上下文）
-- 包含：task metadata, recent_cards, chat_history summary

-- Step 2: Create index for efficient querying
CREATE INDEX IF NOT EXISTS idx_messages_task_id_type
  ON messages(task_id, message_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_referenced_cards
  ON messages USING GIN (referenced_card_ids);

-- Step 3: Add constraints to ensure data integrity
ALTER TABLE messages
  ADD CONSTRAINT check_message_type
  CHECK (message_type IN ('user_query', 'ai_reply', 'system_notification'));

-- Step 4: Create view for filtering chat messages by type
CREATE OR REPLACE VIEW chat_messages_v AS
SELECT
  id,
  task_id,
  role,
  type,
  content,
  message_type,
  agent_config,
  referenced_card_ids,
  parameter_updates,
  context_snapshot,
  created_at,
  updated_at
FROM messages
WHERE message_type IN ('user_query', 'ai_reply', 'system_notification')
ORDER BY created_at DESC;

-- Step 5: Optional - Create a procedure to auto-clean old chat history (beyond 90 days)
-- This is for production use, not enabled by default
-- CREATE OR REPLACE FUNCTION cleanup_old_chats()
-- RETURNS void AS $$
-- BEGIN
--   DELETE FROM messages
--   WHERE message_type IN ('user_query', 'ai_reply', 'system_notification')
--   AND created_at < NOW() - INTERVAL '90 days';
-- END;
-- $$ LANGUAGE plpgsql;

GRANT ALL PRIVILEGES ON chat_messages_v TO junkfilter;
