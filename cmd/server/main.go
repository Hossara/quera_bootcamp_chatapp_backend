package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/auth"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/config"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/handler"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/middleware"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	client, err := ent.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// Run database migrations
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Failed creating schema resources: %v", err)
	}

	// Initialize auth service
	authService := auth.NewAuthService(cfg.Auth.PasetoKey, cfg.Auth.TokenExpiration)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(client, authService)
	userHandler := handler.NewUserHandler(client, authService)
	chatHandler := handler.NewChatHandler(client)
	messageHandler := handler.NewMessageHandler(client)
	wsHandler := handler.NewWebSocketHandler(client, authService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Chat App Backend",
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
	}))

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "chat-app-backend",
		})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Auth routes (public)
	authRoutes := v1.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)

	// Protected routes
	authMiddleware := middleware.AuthMiddleware(authService)

	// Auth protected routes
	authRoutes.Get("/me", authMiddleware, authHandler.GetMe)

	// User routes
	userRoutes := v1.Group("/users", authMiddleware)
	userRoutes.Get("/", userHandler.ListUsers)
	userRoutes.Get("/:id", userHandler.GetUser)
	userRoutes.Put("/:id", userHandler.UpdateUser)
	userRoutes.Delete("/:id", userHandler.DeleteUser)
	userRoutes.Post("/last-seen", userHandler.UpdateLastSeen)

	// Chat routes
	chatRoutes := v1.Group("/chats", authMiddleware)
	chatRoutes.Post("/", chatHandler.CreateChat)
	chatRoutes.Get("/", chatHandler.ListChats)
	chatRoutes.Get("/:id", chatHandler.GetChat)
	chatRoutes.Put("/:id", chatHandler.UpdateChat)
	chatRoutes.Delete("/:id", chatHandler.DeleteChat)
	chatRoutes.Post("/:id/members", chatHandler.AddMembers)
	chatRoutes.Delete("/:id/members/:memberId", chatHandler.RemoveMember)

	// Message routes
	messageRoutes := v1.Group("/messages", authMiddleware)
	messageRoutes.Post("/", messageHandler.SendMessage)
	messageRoutes.Get("/:id", messageHandler.GetMessage)
	messageRoutes.Get("/chat/:chatId", messageHandler.ListMessages)
	messageRoutes.Put("/:id", messageHandler.UpdateMessage)
	messageRoutes.Delete("/:id", messageHandler.DeleteMessage)

	// WebSocket route
	app.Get("/ws", wsHandler.HandleWebSocket())
	app.Get("/ws/health", wsHandler.HealthCheck)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func customErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
