package handler

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/auth"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/model"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/service"
	f "github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/fiber"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	userService *service.UserService
	authService *auth.Service
}

func NewAuthHandler(client *ent.Client, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		userService: service.NewUserService(client, authService),
		authService: authService,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	req := new(model.RegisterRequest)
	if err := f.ParseRequestBody(c, req); err != nil {
		return f.RespondError(c, fiber.StatusBadRequest, err.Message, err.Errors)
	}

	// Validate username
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "username must be between 3 and 50 characters",
		})
	}

	// Validate password
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "password must be at least 6 characters",
		})
	}

	// Check if user exists
	exists, err := h.userService.UserExists(context.Background(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to check user existence",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(model.ErrorResponse{
			Error: "username already exists",
		})
	}

	// Create user
	newUser, err := h.userService.CreateUser(context.Background(), req.Username, req.Password, req.DisplayName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to create user",
		})
	}

	// Generate token
	token, err := h.userService.CreateToken(newUser.ID, newUser.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.AuthResponse{
		Token: token,
		User: model.UserProfile{
			ID:          newUser.ID,
			Username:    newUser.Username,
			DisplayName: newUser.DisplayName,
			CreatedAt:   newUser.CreatedAt,
			LastSeen:    newUser.LastSeen,
		},
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	req := new(model.LoginRequest)
	if err := f.ParseRequestBody(c, req); err != nil {
		return f.RespondError(c, fiber.StatusBadRequest, err.Message, err.Errors)
	}

	// Find user
	u, err := h.userService.GetUserByUsername(context.Background(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Error: "invalid credentials",
		})
	}

	// Verify password
	if err := h.userService.VerifyPassword(u.Password, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Error: "invalid credentials",
		})
	}

	// Generate token
	token, err := h.userService.CreateToken(u.ID, u.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to generate token",
		})
	}

	return c.JSON(model.AuthResponse{
		Token: token,
		User: model.UserProfile{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			CreatedAt:   u.CreatedAt,
			LastSeen:    u.LastSeen,
		},
	})
}

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	u, err := h.userService.GetUserByID(context.Background(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
			Error: "user not found",
		})
	}

	return c.JSON(model.UserProfile{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
		LastSeen:    u.LastSeen,
	})
}

func convertNullTimeToPointer(nt sql.NullTime) *interface{} {
	if !nt.Valid {
		return nil
	}
	var t interface{} = nt.Time
	return &t
}
