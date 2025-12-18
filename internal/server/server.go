package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/config"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/auth"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/handler"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/middleware"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// Server represents the application server
type Server struct {
	app         *fiber.App
	config      *config.Config
	client      *ent.Client
	authService *auth.Service
}

// New creates a new server instance
func New(cfg *config.Config, client *ent.Client) (*Server, error) {
	// Initialize auth service
	authService := auth.NewAuthService(cfg.Auth.PasetoKey, cfg.Auth.TokenExpiration)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Chat App Backend",
		ServerHeader: "Fiber",
		ErrorHandler: customErrorHandler,
	})

	server := &Server{
		app:         app,
		config:      cfg,
		client:      client,
		authService: authService,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	return server, nil
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	s.app.Use(recover.New())
	s.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
	}))
}

// setupRoutes configures all application routes
func (s *Server) setupRoutes() {
	// Initialize handlers
	authHandler := handler.NewAuthHandler(s.client, s.authService)
	userHandler := handler.NewUserHandler(s.client, s.authService)
	chatHandler := handler.NewChatHandler(s.client)
	messageHandler := handler.NewMessageHandler(s.client)
	wsHandler := handler.NewWebSocketHandler(s.client, s.authService)

	// Health check
	s.app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "chat-app-backend",
		})
	})

	// API v1 routes
	v1 := s.app.Group("/api/v1")

	// Auth routes (public)
	authRoutes := v1.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)

	// Protected routes
	authMiddleware := middleware.AuthMiddleware(s.authService)

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
	s.app.Get("/ws", wsHandler.HandleWebSocket())
	s.app.Get("/ws/health", wsHandler.HealthCheck)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Environment: %s", s.config.Server.Environment)

	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}

// Close closes database connections and other resources
func (s *Server) Close() error {
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}
	return nil
}
