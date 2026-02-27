-- ============================================================
-- Junk Filter 完整数据库 Schema
-- ============================================================

-- 1. 博主表 (Bloggers)
CREATE TABLE IF NOT EXISTS bloggers (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  bio TEXT,
  location VARCHAR(255),
  avatar VARCHAR(1000),
  rss_feed VARCHAR(500) UNIQUE,
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'paused', 'blocked')),
  filter_rate DECIMAL(5,2) DEFAULT 0 CHECK (filter_rate >= 0 AND filter_rate <= 100),
  total_articles INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blogger_status ON bloggers(status);
CREATE INDEX IF NOT EXISTS idx_blogger_rss ON bloggers(rss_feed);

-- 2. RSS 源表 (Feeds)
CREATE TABLE IF NOT EXISTS feeds (
  id SERIAL PRIMARY KEY,
  url VARCHAR(500) UNIQUE NOT NULL,
  name VARCHAR(255),
  update_frequency VARCHAR(20) DEFAULT 'hourly' CHECK (update_frequency IN ('hourly', '30min', '2hours')),
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'paused', 'failed')),
  last_fetch_time TIMESTAMP,
  error_count INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_status ON feeds(status);
CREATE INDEX IF NOT EXISTS idx_feed_last_fetch ON feeds(last_fetch_time);

-- 3. 内容表 (Contents)
CREATE TABLE IF NOT EXISTS contents (
  id SERIAL PRIMARY KEY,
  blogger_id INTEGER NOT NULL REFERENCES bloggers(id) ON DELETE CASCADE,
  title VARCHAR(500) NOT NULL,
  summary TEXT,
  original_url VARCHAR(1000) UNIQUE,
  published_at TIMESTAMP NOT NULL,

  -- AI 评分
  quality_score DECIMAL(5,2) CHECK (quality_score >= 0 AND quality_score <= 100),
  relevance_score DECIMAL(5,2) CHECK (relevance_score >= 0 AND relevance_score <= 100),
  ai_decision VARCHAR(20) DEFAULT 'pending' CHECK (ai_decision IN ('pending', 'approved', 'rejected', 'review')),

  -- 原始 TrueSignal 字段兼容
  innovation_score INTEGER,
  depth_score INTEGER,
  decision VARCHAR(20),
  tldr TEXT,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_blogger ON contents(blogger_id);
CREATE INDEX IF NOT EXISTS idx_content_decision ON contents(ai_decision);
CREATE INDEX IF NOT EXISTS idx_content_published ON contents(published_at);
CREATE INDEX IF NOT EXISTS idx_content_quality ON contents(quality_score);

-- 4. 任务表 (Tasks)
CREATE TABLE IF NOT EXISTS tasks (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,

  -- 时间调度
  schedule VARCHAR(100) NOT NULL,  -- cron 格式: "0 9 * * *"
  time_display VARCHAR(20),        -- 显示用："09:00 AM"
  enabled BOOLEAN DEFAULT true,

  -- 任务配置
  type VARCHAR(20) NOT NULL CHECK (type IN ('summary', 'filter', 'monitor')),
  config JSONB DEFAULT '{}',       -- 存储 platforms, keywords, thresholds 等

  -- 执行统计
  last_executed_at TIMESTAMP,
  next_execute_at TIMESTAMP,
  execution_count INTEGER DEFAULT 0,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_enabled ON tasks(enabled);
CREATE INDEX IF NOT EXISTS idx_task_next_execute ON tasks(next_execute_at);

-- 5. AI 配置表 (AI Config)
CREATE TABLE IF NOT EXISTS ai_config (
  id SERIAL PRIMARY KEY,
  default_model VARCHAR(255) DEFAULT 'GPT-4o',
  api_key TEXT NOT NULL,
  temperature DECIMAL(3,2) DEFAULT 0.7 CHECK (temperature >= 0 AND temperature <= 1),
  max_tokens INTEGER DEFAULT 2048,
  batch_size INTEGER DEFAULT 5,
  retry_count INTEGER DEFAULT 3,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- 6. 内容统计表 (Content Stats - 用于图表)
CREATE TABLE IF NOT EXISTS content_stats (
  id SERIAL PRIMARY KEY,
  blogger_id INTEGER NOT NULL REFERENCES bloggers(id) ON DELETE CASCADE,
  stat_date DATE NOT NULL,

  approved_count INTEGER DEFAULT 0,
  rejected_count INTEGER DEFAULT 0,
  review_count INTEGER DEFAULT 0,
  total_count INTEGER DEFAULT 0,

  avg_quality_score DECIMAL(5,2),
  avg_relevance_score DECIMAL(5,2),

  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE (blogger_id, stat_date),
  INDEX idx_stats_blogger (blogger_id),
  INDEX idx_stats_date (stat_date)
);

-- 7. 任务执行日志表 (Task Execution Log)
CREATE TABLE IF NOT EXISTS task_logs (
  id SERIAL PRIMARY KEY,
  task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,

  status VARCHAR(20) CHECK (status IN ('success', 'failed', 'running')),
  executed_at TIMESTAMP DEFAULT NOW(),
  completed_at TIMESTAMP,

  items_processed INTEGER DEFAULT 0,
  items_created INTEGER DEFAULT 0,
  error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_task_logs_task ON task_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_task_logs_status ON task_logs(status);

-- 8. 操作审计日志 (Audit Log)
CREATE TABLE IF NOT EXISTS audit_logs (
  id SERIAL PRIMARY KEY,
  action VARCHAR(100) NOT NULL,
  entity_type VARCHAR(50),        -- 'blogger', 'task', 'feed', etc
  entity_id INTEGER,
  old_data JSONB,
  new_data JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  created_by VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);

-- ============================================================
-- 初始化数据
-- ============================================================

-- 插入默认 AI 配置
INSERT INTO ai_config (default_model, api_key, temperature, max_tokens)
VALUES ('GPT-4o', 'sk-placeholder', 0.7, 2048)
ON CONFLICT DO NOTHING;

-- 创建索引用于性能优化
CREATE INDEX IF NOT EXISTS idx_contents_blogger_decision ON contents(blogger_id, ai_decision);
CREATE INDEX IF NOT EXISTS idx_contents_created_at ON contents(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bloggers_created_at ON bloggers(created_at DESC);

-- 权限设置（如果需要）
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO truesignal;

-- ============================================================
-- 视图定义（方便查询）
-- ============================================================

-- 博主统计视图
CREATE OR REPLACE VIEW blogger_stats AS
SELECT
  b.id,
  b.name,
  b.rss_feed,
  COUNT(c.id) as total_articles,
  SUM(CASE WHEN c.ai_decision = 'approved' THEN 1 ELSE 0 END) as approved_count,
  SUM(CASE WHEN c.ai_decision = 'rejected' THEN 1 ELSE 0 END) as rejected_count,
  AVG(c.quality_score) as avg_quality,
  AVG(c.relevance_score) as avg_relevance,
  MAX(c.published_at) as latest_article
FROM bloggers b
LEFT JOIN contents c ON b.id = c.blogger_id
GROUP BY b.id, b.name, b.rss_feed;

-- 任务执行统计视图
CREATE OR REPLACE VIEW task_execution_stats AS
SELECT
  t.id,
  t.name,
  COUNT(tl.id) as total_executions,
  SUM(CASE WHEN tl.status = 'success' THEN 1 ELSE 0 END) as successful_executions,
  SUM(tl.items_processed) as total_items_processed,
  MAX(tl.executed_at) as last_executed_at
FROM tasks t
LEFT JOIN task_logs tl ON t.id = tl.task_id
GROUP BY t.id, t.name;
