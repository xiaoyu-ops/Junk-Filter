package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junkfilter/backend-go/internal/config"
	"github.com/junkfilter/backend-go/internal/domain"
	"github.com/junkfilter/backend-go/internal/service"
)

// App 是应用上下文（简化版本）
// 相比之前的 AppContext，更清晰和模块化
type App struct {
	factory    *service.Factory
	rssFetcher domain.RSSFetcher
	stopChan   chan struct{}
}

func main() {
	log.Println("\n════════════════════════════════════════════════")
	log.Println("    JunkFilter Backend - Go Service")
	log.Println("════════════════════════════════════════════════\n")

	// Step 1: 加载配置
	log.Println("[1/4] Loading configuration...")
	cfg := config.Load()
	log.Printf("✓ Configuration loaded from: config.yaml + environment variables\n")

	// Step 2: 初始化服务工厂（依赖注入的核心）
	log.Println("[2/4] Initializing service factory...")
	factory, err := service.NewFactory(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize factory: %v", err)
	}
	defer factory.Close()
	log.Println("✓ Service factory initialized\n")

	// Step 3: 创建 RSS 抓取器
	log.Println("[3/4] Creating RSS fetcher...")
	rssFetcher, err := factory.CreateRSSFetcher()
	if err != nil {
		log.Fatalf("Failed to create RSS fetcher: %v", err)
	}
	log.Println("✓ RSS fetcher created\n")

	// Step 4: 启动应用
	log.Println("[4/4] Starting application...")
	app := &App{
		factory:    factory,
		rssFetcher: rssFetcher,
		stopChan:   make(chan struct{}),
	}

	if err := app.start(cfg); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// 保持主程序运行
	select {}
}

// start 启动应用：RSS 服务 + HTTP 服务器
func (app *App) start(cfg *config.Config) error {
	// 启动 RSS 服务（后台）
	fetchInterval := cfg.GetFetchInterval()
	go func() {
		if err := app.rssFetcher.Start(context.Background(), fetchInterval); err != nil {
			log.Printf("Error starting RSS service: %v", err)
		}
	}()
	defer app.rssFetcher.Stop()

	// 启动 HTTP 服务（后台）
	go app.startHTTPServer(cfg)

	return nil
}

// startHTTPServer 启动 HTTP 服务器
func (app *App) startHTTPServer(cfg *config.Config) {
	router := gin.Default()

	// CORS 中间件
	router.Use(corsMiddleware())

	// 注册 handlers
	app.registerHandlers(router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port)
	log.Printf("✓ HTTP Server listening on %s\n", addr)
	log.Println("\n════════════════════════════════════════════════")
	log.Println("    Application is running. Press Ctrl+C to stop.")
	log.Println("════════════════════════════════════════════════\n")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// registerHandlers 注册所有 HTTP handlers
// 注：实际集成时需要根据现有 handlers 的实现进行调整
func (app *App) registerHandlers(router *gin.Engine) {
	// 从工厂获取仓储和服务
	// sourceRepo := app.factory.SourceRepo()
	// contentRepo := app.factory.ContentRepo()
	// publisher := app.factory.Publisher()

	// TODO: 根据实际的 handlers 实现来注册路由
	// 这里是示例框架，实际需要：
	// 1. 检查 handlers 包中的函数签名
	// 2. 做适当的类型转换和适配
	// 3. 确保仓储和服务正确注入

	log.Println("✓ HTTP handlers registered (routes to be implemented)")
}

// corsMiddleware 返回 CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
