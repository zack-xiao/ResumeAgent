package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"resume-agent/config"
	"resume-agent/handler"
	"resume-agent/persona"
	"resume-agent/service"
)

func main() {
	// 加载 .env 文件
	godotenv.Load("../.env")

	// 加载配置
	cfg := config.Load()

	if cfg.DeepSeekAPIKey == "" {
		fmt.Println("错误: 请设置 DEEPSEEK_API_KEY 环境变量")
		fmt.Println("示例: export DEEPSEEK_API_KEY=sk-xxxxxx")
		os.Exit(1)
	}

	// 加载人物设定
	personaLoader := persona.NewLoader(cfg.PersonaPath)
	personaContent, err := personaLoader.Load()
	if err != nil {
		log.Fatalf("加载人物设定失败: %v", err)
	}

	// 创建聊天服务
	chatService := service.NewChatService(service.Config{
		APIKey:  cfg.DeepSeekAPIKey,
		Model:   cfg.DeepSeekModel,
		Persona: personaContent,
	})

	// 创建处理器
	chatHandler := handler.NewChatHandler(chatService, personaLoader, cfg)

	// 打印配置信息
	if cfg.AccessPassword != "" {
		fmt.Println("访问密码: 已启用")
	} else {
		fmt.Println("访问密码: 未启用")
	}

	// 设置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件服务 - 使用绝对路径
	frontendPath, _ := filepath.Abs("../frontend")
	dataPath, _ := filepath.Abs("../data")
	r.Static("/static", frontendPath)
	r.Static("/data", dataPath)

	// 前端页面和简历路径
	frontendPathAbs, _ := filepath.Abs("../frontend")
	dataPathAbs, _ := filepath.Abs("../data")

	// API 路由
	api := r.Group("/api")
	{
		api.GET("/init", chatHandler.InitHandler)
		api.POST("/verify", chatHandler.VerifyHandler)
		api.POST("/chat", chatHandler.ChatHandler)
		api.POST("/chat/stream", chatHandler.StreamHandler)
		api.POST("/reload", chatHandler.ReloadHandler)
		api.GET("/resume", func(c *gin.Context) {
			c.File(filepath.Join(dataPathAbs, "肖正烁的简历.pdf"))
		})
	}

	// 前端页面
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(frontendPathAbs, "index.html"))
	})

	addr := ":" + cfg.Port
	fmt.Printf("启动服务: http://localhost:%s\n", cfg.Port)
	fmt.Println("按 Ctrl+C 停止服务")

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
