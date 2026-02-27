-- TrueSignal PostgreSQL Schema (Simplified for Demo)

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 表 1: sources (RSS 源)
CREATE TABLE IF NOT EXISTS sources (
  id BIGSERIAL PRIMARY KEY,
  platform VARCHAR(50) DEFAULT 'blog',
  url VARCHAR(500) UNIQUE NOT NULL,
  author_name VARCHAR(200),
  author_id VARCHAR(200),
  priority INT DEFAULT 5,
  last_fetch_time TIMESTAMP,
  fetch_interval_seconds INT DEFAULT 3600,
  enabled BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sources_enabled_priority ON sources(enabled, priority);

-- 表 2: content (文章主体)
CREATE TABLE IF NOT EXISTS content (
  id BIGSERIAL PRIMARY KEY,
  task_id UUID UNIQUE DEFAULT uuid_generate_v4(),
  source_id BIGINT REFERENCES sources(id) ON DELETE CASCADE,
  platform VARCHAR(50) DEFAULT 'blog',
  author_name VARCHAR(200),
  title VARCHAR(500),
  original_url VARCHAR(500) UNIQUE,
  content_hash VARCHAR(64) UNIQUE,
  clean_content TEXT,
  published_at TIMESTAMP,
  ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  status VARCHAR(50) DEFAULT 'PENDING',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_content_status ON content(status);
CREATE INDEX idx_content_published_at ON content(published_at DESC);
CREATE INDEX idx_content_author ON content(author_name);
CREATE INDEX idx_content_task_id ON content(task_id);

-- 表 3: evaluation (评估结果)
CREATE TABLE IF NOT EXISTS evaluation (
  id BIGSERIAL PRIMARY KEY,
  content_id BIGINT REFERENCES content(id) ON DELETE CASCADE UNIQUE,
  task_id UUID REFERENCES content(task_id),
  innovation_score INT,
  depth_score INT,
  decision VARCHAR(50),
  reasoning TEXT,
  tldr TEXT,
  key_concepts TEXT[],
  evaluated_at TIMESTAMP,
  evaluator_version VARCHAR(50),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_evaluation_decision ON evaluation(decision);
CREATE INDEX idx_evaluation_innovation_score ON evaluation(innovation_score DESC);
CREATE INDEX idx_evaluation_depth_score ON evaluation(depth_score DESC);
CREATE INDEX idx_evaluation_task_id ON evaluation(task_id);

-- 表 4: user_subscription (订阅关系) - Demo 版本简化
CREATE TABLE IF NOT EXISTS user_subscription (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID DEFAULT uuid_generate_v4(),
  source_id BIGINT REFERENCES sources(id) ON DELETE CASCADE,
  subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  min_innovation_score INT DEFAULT 5,
  min_depth_score INT DEFAULT 4
);

CREATE INDEX idx_user_subscription_user_id ON user_subscription(user_id);

-- 表 5: status_log (状态机日志)
CREATE TABLE IF NOT EXISTS status_log (
  id BIGSERIAL PRIMARY KEY,
  content_id BIGINT REFERENCES content(id) ON DELETE CASCADE,
  task_id UUID,
  from_status VARCHAR(50),
  to_status VARCHAR(50),
  reason TEXT,
  logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_status_log_task_id ON status_log(task_id);
CREATE INDEX idx_status_log_logged_at ON status_log(logged_at DESC);

-- 表 6: messages (聊天消息 - Phase 5 新增)
CREATE TABLE IF NOT EXISTS messages (
  id BIGSERIAL PRIMARY KEY,
  task_id BIGINT REFERENCES sources(id) ON DELETE CASCADE,
  role VARCHAR(20) NOT NULL,  -- 'user', 'ai'
  type VARCHAR(20) DEFAULT 'text',  -- 'text', 'execution'
  content TEXT NOT NULL,
  metadata JSONB,  -- 存储额外信息如 execution 结果
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_task_id ON messages(task_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_role ON messages(role);

-- Demo 初始数据：添加几个测试源
INSERT INTO sources (url, author_name, priority, enabled) VALUES
  ('https://feeds.arstechnica.com/arstechnica/index', 'Ars Technica', 8, TRUE),
  ('https://news.ycombinator.com/rss', 'Hacker News', 9, TRUE),
  ('https://feeds.medium.com/tag/technology/latest', 'Medium Tech', 7, TRUE)
ON CONFLICT (url) DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO truesignal;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO truesignal;
