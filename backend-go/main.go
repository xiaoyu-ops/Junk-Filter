package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"

	"github.com/junkfilter/backend-go/handlers"
	"github.com/junkfilter/backend-go/repositories"
	"github.com/junkfilter/backend-go/services"
)

// Config 应用配置
type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
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
		WorkerCount   int    `yaml:"worker_count"`
		Timeout       string `yaml:"timeout"`
		RetryMax      int    `yaml:"retry_max"`
		FetchInterval string `yaml:"fetch_interval"`
	} `yaml:"ingestion"`
}

// AppContext 应用上下文
type AppContext struct {
	DB             *sql.DB
	Redis          *redis.Client
	Config         *Config
	RSSService     *services.RSSService
	SourceRepo     *repositories.SourceRepository
	ContentRepo    *repositories.ContentRepository
	EvaluationRepo *repositories.EvaluationRepository
	MessageRepo    *repositories.MessageRepository
}

var appCtx *AppContext

func main() {
	// 加载配置
	cfg := loadConfig()
	log.Println("✓ Configuration loaded")

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("✓ Database connected")

	// 初始化 Redis
	rdb := initRedis(cfg)
	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✓ Redis connected")

	// 初始化 repositories
	sourceRepo := repositories.NewSourceRepository(db)
	contentRepo := repositories.NewContentRepository(db)
	evaluationRepo := repositories.NewEvaluationRepository(db)
	messageRepo := repositories.NewMessageRepository(db)

	// 初始化 services
	contentService := services.NewContentService(rdb)
	parseFetchTimeout := 10 * time.Second
	if cfg.Ingestion.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Ingestion.Timeout); err == nil {
			parseFetchTimeout = d
		}
	}

	rssService := services.NewRSSService(
		sourceRepo,
		contentRepo,
		rdb,
		contentService,
		cfg.Ingestion.WorkerCount,
		parseFetchTimeout,
		cfg.Ingestion.RetryMax,
	)

	// 保存到全局上下文
	appCtx = &AppContext{
		DB:             db,
		Redis:          rdb,
		Config:         cfg,
		RSSService:     rssService,
		SourceRepo:     sourceRepo,
		ContentRepo:    contentRepo,
		EvaluationRepo: evaluationRepo,
		MessageRepo:    messageRepo,
	}

	log.Println("\n========== JunkFilter Backend ==========")
	log.Printf("Database: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	log.Printf("Redis: %s:%d\n", cfg.Redis.Host, cfg.Redis.Port)
	log.Printf("Server: listening on :%d\n", cfg.Server.Port)
	log.Println("========================================\n")

	// 启动 RSS 服务（异步，不会阻塞）
	fetchInterval := 1 * time.Hour
	if cfg.Ingestion.FetchInterval != "" {
		if d, err := time.ParseDuration(cfg.Ingestion.FetchInterval); err == nil {
			fetchInterval = d
		}
	}

	// 在后台启动 RSS 服务（不要在主线程等待）
	go func() {
		if err := rssService.Start(context.Background(), fetchInterval); err != nil {
			log.Printf("Error starting RSS service: %v", err)
		}
	}()
	defer rssService.Stop()

	// 启动 HTTP 服务（在后台异步运行）
	log.Println("DEBUG: About to call startServer()")
	go startServer(cfg.Server.Port)
	log.Println("DEBUG: startServer() started in background")

	// 保持主程序运行
	select {}
}

func loadConfig() *Config {
	cfg := &Config{}

	// Default configuration
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "junkfilter"
	cfg.Database.Password = "junkfilter123"
	cfg.Database.DBName = "junkfilter"
	cfg.Database.SSLMode = "disable"
	cfg.Redis.Host = "localhost"
	cfg.Redis.Port = 6379
	cfg.Redis.DB = 0
	cfg.Server.Port = 8080
	cfg.Ingestion.WorkerCount = 5
	cfg.Ingestion.Timeout = "10s"
	cfg.Ingestion.RetryMax = 3
	cfg.Ingestion.FetchInterval = "1h"

	// 尝试从 config.yaml 读取
	if data, err := os.ReadFile("config.yaml"); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			log.Printf("Warning: Failed to parse config.yaml: %v\n", err)
		}
	}

	// Environment variable overrides
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Database.DBName = dbname
	}

	return cfg
}

func initDatabase(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 配置连接池
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	return db, nil
}

func initRedis(cfg *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	})
}

func startServer(port int) {
	router := gin.Default()

	// 添加 CORS 中间件
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 注册 handlers
	sourceHandler := handlers.NewSourceHandler(appCtx.SourceRepo, appCtx.RSSService)
	contentHandler := handlers.NewContentHandler(appCtx.ContentRepo, appCtx.EvaluationRepo)
	evaluationHandler := handlers.NewEvaluationHandler(appCtx.EvaluationRepo)
	messageHandler := handlers.NewMessageHandler(appCtx.MessageRepo)
	chatHandler := handlers.NewChatHandler(appCtx.MessageRepo, "http://localhost:8081")
	taskChatHandler := handlers.NewTaskChatHandler(
		appCtx.MessageRepo,
		appCtx.SourceRepo,
		appCtx.EvaluationRepo,
		"http://localhost:8081",
	)

	// 注册路由
	handlers.RegisterSourceRoutes(router, sourceHandler)
	handlers.RegisterContentRoutes(router, contentHandler)
	handlers.RegisterEvaluationRoutes(router, evaluationHandler)
	handlers.RegisterMessageRoutes(router, messageHandler)
	handlers.RegisterChatRoutes(router, chatHandler)
	handlers.RegisterTaskChatRoutes(router, taskChatHandler)

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("DEBUG: About to start server on %s\n", addr)
	log.Printf("DEBUG: router.Run() is blocking call\n")
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
	log.Printf("DEBUG: Server exited (this should not print)\n")
}

