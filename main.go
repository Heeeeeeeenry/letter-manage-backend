package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"letter-manage-backend/config"
	"letter-manage-backend/controller"
	"letter-manage-backend/dao"
	"letter-manage-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("load config: %v", err)
	}
	// Start auto-reload every 1 minute
	config.StartAutoReload("config.yaml")

	cfg := config.Get()
	gin.SetMode(cfg.Server.Mode)

	// Init DB
	if err := dao.InitDB(); err != nil {
		log.Fatalf("init db: %v", err)
	}
	log.Println("Database connected")

	// Auto migrate
	if err := dao.AutoMigrate(); err != nil {
		log.Printf("auto migrate warning: %v", err)
	}

	// Setup router
	r := gin.Default()

	// CORS
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// Static media files
	r.Static("/media", cfg.Media.Root)

	// Static frontend files
	r.Static("/assets", "./assets")
	r.Static("/css", "./css")
	r.Static("/js", "./js")
	r.StaticFile("/", "./pages/index.html")

	// API routes
	api := r.Group("/api")

	// Auth (no auth middleware)
	api.POST("/auth/", controller.AuthController)

	// Config (optional auth for menus)
	api.POST("/config/", controller.ConfigController)

	// Tool routes (no auth required for tools)
	toolGroup := api.Group("/tool")
	toolGroup.POST("/", controller.ToolController)
	toolGroup.POST("/time_diff/", controller.ToolTimeDiff)
	toolGroup.POST("/time_add/", controller.ToolTimeAdd)
	toolGroup.POST("/holiday_check/", controller.ToolHolidayCheck)
	toolGroup.POST("/workdays_calculate/", controller.ToolWorkdaysCalculate)
	toolGroup.POST("/workdays_add/", controller.ToolWorkdaysAdd)
	toolGroup.POST("/month_calendar/", controller.ToolMonthCalendar)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired())

	protected.POST("/letter/", controller.LetterController)
	protected.POST("/setting/", controller.SettingController)
	protected.POST("/llm/", controller.LLMController)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Background cleanup of expired sessions
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := dao.CleanExpiredSessions(); err != nil {
				log.Printf("clean sessions error: %v", err)
			}
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
