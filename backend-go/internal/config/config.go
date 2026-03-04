package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置（统一管理）
type Config struct {
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		DBName       string `yaml:"dbname"`
		SSLMode      string `yaml:"sslmode"`
		MaxOpenConns int    `yaml:"max_open_conns"`  // P0: 连接池优化
		MaxIdleConns int    `yaml:"max_idle_conns"`  // P0: 连接池优化
	} `yaml:"database"`
	Redis struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		DB   int    `yaml:"db"`
	} `yaml:"redis"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Ingestion struct {
		WorkerCount   int    `yaml:"worker_count"`    // P0: 优化值 20
		Timeout       string `yaml:"timeout"`         // P0: 优化值 30s
		RetryMax      int    `yaml:"retry_max"`
		FetchInterval string `yaml:"fetch_interval"` // P0: 优化值 30m
	} `yaml:"ingestion"`
}

// Load 加载配置（YAML + 环境变量覆盖）
func Load() *Config {
	cfg := &Config{}

	// Step 1: 设置默认值（P0 优化后的默认值）
	cfg.setDefaults()

	// Step 2: 从 config.yaml 读取
	if data, err := os.ReadFile("config.yaml"); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			log.Printf("Warning: Failed to parse config.yaml: %v\n", err)
		}
	}

	// Step 3: 环境变量覆盖
	cfg.applyEnvironmentOverrides()

	return cfg
}

// setDefaults 设置默认值（含 P0 优化值）
func (c *Config) setDefaults() {
	c.Database.Host = "localhost"
	c.Database.Port = 5432
	c.Database.User = "junkfilter"
	c.Database.Password = "junkfilter123"
	c.Database.DBName = "junkfilter"
	c.Database.SSLMode = "disable"
	c.Database.MaxOpenConns = 50   // P0: 优化值
	c.Database.MaxIdleConns = 10   // P0: 优化值

	c.Redis.Host = "localhost"
	c.Redis.Port = 6379
	c.Redis.DB = 0

	c.Server.Port = 8080

	c.Ingestion.WorkerCount = 20       // P0: 优化值（从 5 改为 20）
	c.Ingestion.Timeout = "30s"        // P0: 优化值（从 10s 改为 30s）
	c.Ingestion.RetryMax = 3
	c.Ingestion.FetchInterval = "30m"  // P0: 优化值（从 1h 改为 30m）
}

// applyEnvironmentOverrides 应用环境变量覆盖
func (c *Config) applyEnvironmentOverrides() {
	// Database
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.User = user
	}
	if pwd := os.Getenv("DB_PASSWORD"); pwd != "" {
		c.Database.Password = pwd
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		c.Database.DBName = dbname
	}

	// Redis
	if host := os.Getenv("REDIS_HOST"); host != "" {
		c.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Redis.Port = p
		}
	}

	// Server
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}

	// Ingestion
	if workers := os.Getenv("INGESTION_WORKERS"); workers != "" {
		if w, err := strconv.Atoi(workers); err == nil {
			c.Ingestion.WorkerCount = w
		}
	}
	if timeout := os.Getenv("INGESTION_TIMEOUT"); timeout != "" {
		c.Ingestion.Timeout = timeout
	}
	if interval := os.Getenv("INGESTION_FETCH_INTERVAL"); interval != "" {
		c.Ingestion.FetchInterval = interval
	}
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.DBName, c.Database.SSLMode,
	)
}

// GetRedisAddr 获取 Redis 地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetFetchTimeout 获取抓取超时时间
func (c *Config) GetFetchTimeout() time.Duration {
	if d, err := time.ParseDuration(c.Ingestion.Timeout); err == nil {
		return d
	}
	return 30 * time.Second
}

// GetFetchInterval 获取抓取间隔
func (c *Config) GetFetchInterval() time.Duration {
	if d, err := time.ParseDuration(c.Ingestion.FetchInterval); err == nil {
		return d
	}
	return 30 * time.Minute
}
