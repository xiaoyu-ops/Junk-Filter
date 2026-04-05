-- Add eval_attempts column to track LLM retry count per content item
ALTER TABLE content ADD COLUMN IF NOT EXISTS eval_attempts INTEGER DEFAULT 0;
