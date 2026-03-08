package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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
		MaxOpenConns int `yaml:"max_open_conns"`  // ← P0: 新增支持
		MaxIdleConns int `yaml:"max_idle_conns"`  // ← P0: 新增支持
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		DB       int    `yaml:"db"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	PythonAPI struct {
		URL string `yaml:"url"`  // P1-4: Python 后端 API URL
	} `yaml:"python_api"`
	CORS struct {
		AllowedOrigins   []string `yaml:"allowed_origins"`
		AllowedMethods   []string `yaml:"allowed_methods"`
		AllowedHeaders   []string `yaml:"allowed_headers"`
		AllowCredentials bool     `yaml:"allow_credentials"`
		MaxAge           int      `yaml:"max_age"`
	} `yaml:"cors"`
	Ingestion struct {
		WorkerCount   int    `yaml:"worker_count"`
		Timeout       string `yaml:"timeout"`
		RetryMax      int    `yaml:"retry_max"`
		FetchInterval string `yaml:"fetch_interval"`
		ProxyURL      string `yaml:"proxy_url"`
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
	ThreadRepo     *repositories.ThreadRepository
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
	threadRepo := repositories.NewThreadRepository(db)

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
		cfg.Ingestion.ProxyURL,
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
		ThreadRepo:     threadRepo,
	}

	log.Println("\n========== JunkFilter Backend ==========")
	log.Printf("Database: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	log.Printf("Redis: %s:%d\n", cfg.Redis.Host, cfg.Redis.Port)
	if cfg.Ingestion.ProxyURL != "" {
		log.Printf("RSS Proxy: %s\n", cfg.Ingestion.ProxyURL)
	} else {
		log.Println("RSS Proxy: disabled (set RSS_PROXY_URL to enable)")
	}
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
	cfg.Redis.Password = ""
	cfg.Server.Port = 8080
	cfg.PythonAPI.URL = "http://localhost:8083"  // P1-4: Python API 默认地址
	cfg.Ingestion.WorkerCount = 5
	cfg.Ingestion.Timeout = "10s"
	cfg.Ingestion.RetryMax = 3
	cfg.Ingestion.FetchInterval = "1h"

	// CORS 默认值 - 严格模式
	cfg.CORS.AllowedOrigins = []string{"http://localhost:5173"}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	cfg.CORS.AllowCredentials = false
	cfg.CORS.MaxAge = 3600

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
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		cfg.Database.SSLMode = sslMode
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Redis.Password = password
	}
	if pythonAPI := os.Getenv("PYTHON_API_URL"); pythonAPI != "" {
		cfg.PythonAPI.URL = strings.TrimSpace(pythonAPI)
	}
	if proxyURL := os.Getenv("RSS_PROXY_URL"); proxyURL != "" {
		cfg.Ingestion.ProxyURL = proxyURL
	}

	// CORS 环境变量覆盖
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		cfg.CORS.AllowedOrigins = strings.Split(origins, ",")
	}
	if methods := os.Getenv("CORS_ALLOWED_METHODS"); methods != "" {
		cfg.CORS.AllowedMethods = strings.Split(methods, ",")
	}
	if headers := os.Getenv("CORS_ALLOWED_HEADERS"); headers != "" {
		cfg.CORS.AllowedHeaders = strings.Split(headers, ",")
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

	// 配置连接池 (P0 优化)
	maxOpenConns := cfg.Database.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 50  // 默认值
	}
	maxIdleConns := cfg.Database.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 10  // 默认值
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	log.Printf("✓ Database pool configured: max_open=%d, max_idle=%d", maxOpenConns, maxIdleConns)

	return db, nil
}

func initRedis(cfg *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

func startServer(port int) {
	router := gin.Default()

	// 添加 CORS 中间件 - 环境变量驱动的严格模式
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查源是否在允许列表中
		isAllowed := false
		for _, allowed := range appCtx.Config.CORS.AllowedOrigins {
			if allowed == "*" || allowed == origin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			if appCtx.Config.CORS.AllowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		}

		// 只允许配置中的方法和头
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(appCtx.Config.CORS.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(appCtx.Config.CORS.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", appCtx.Config.CORS.MaxAge))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 注册 handlers
	sourceHandler := handlers.NewSourceHandler(appCtx.SourceRepo, appCtx.RSSService)
	contentHandler := handlers.NewContentHandler(appCtx.ContentRepo, appCtx.EvaluationRepo, appCtx.SourceRepo, appCtx.DB)
	evaluationHandler := handlers.NewEvaluationHandler(appCtx.EvaluationRepo)
	messageHandler := handlers.NewMessageHandler(appCtx.MessageRepo)
	taskChatHandler := handlers.NewTaskChatHandler(
		appCtx.MessageRepo,
		appCtx.SourceRepo,
		appCtx.EvaluationRepo,
		appCtx.Config.PythonAPI.URL,
	)
	aiTaskHandler := handlers.NewAITaskHandler(appCtx.SourceRepo, appCtx.Config.PythonAPI.URL)
	configHandler := handlers.NewConfigHandler(appCtx.DB)

	// 注册路由
	handlers.RegisterSourceRoutes(router, sourceHandler)
	handlers.RegisterContentRoutes(router, contentHandler)
	handlers.RegisterEvaluationRoutes(router, evaluationHandler)
	handlers.RegisterMessageRoutes(router, messageHandler)
	handlers.RegisterTaskChatRoutes(router, taskChatHandler)
	handlers.RegisterAITaskRoutes(router, aiTaskHandler)
	handlers.RegisterConfigRoutes(router, configHandler)

	threadHandler := handlers.NewThreadHandler(appCtx.ThreadRepo, appCtx.MessageRepo)
	handlers.RegisterThreadRoutes(router, threadHandler)

	// 内容搜索路由（需要注入 db 到 context）
	router.GET("/api/search", func(c *gin.Context) {
		c.Set("db", appCtx.DB)
		handlers.SearchContent(c)
	})

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("✓ Server starting on %s\n", addr)

	// 创建自定义 TCP 监听器（支持端口重用）
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	if err := router.RunListener(listener); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

