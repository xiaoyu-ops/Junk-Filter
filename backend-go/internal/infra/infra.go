package infra

import (
	"database/sql"
	"log"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/junkfilter/backend-go/internal/config"
)

// Database 封装数据库初始化和连接
type Database struct {
	conn *sql.DB
}

// NewDatabase 创建并初始化数据库连接
func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// P0: 配置连接池（参数驱动）
	maxOpenConns := cfg.Database.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 50
	}
	maxIdleConns := cfg.Database.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 10
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	log.Printf("✓ Database initialized: %s:%d/%s (pool: max_open=%d, max_idle=%d)",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName,
		maxOpenConns, maxIdleConns)

	return &Database{conn: db}, nil
}

// Conn 获取数据库连接
func (d *Database) Conn() *sql.DB {
	return d.conn
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// ============================================================================

// Redis 封装 Redis 客户端
type Redis struct {
	client *redis.Client
}

// NewRedis 创建并初始化 Redis 客户端
func NewRedis(cfg *config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.GetRedisAddr(),
		DB:   cfg.Redis.DB,
	})

	// 测试连接
	if err := client.Ping(client.Context()).Err(); err != nil {
		return nil, err
	}

	log.Printf("✓ Redis connected: %s (db=%d)", cfg.GetRedisAddr(), cfg.Redis.DB)

	return &Redis{client: client}, nil
}

// Client 获取 Redis 客户端
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Close 关闭 Redis 连接
func (r *Redis) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
