package handler

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/auth"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/model"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/user"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	client      *ent.Client
	authService *auth.AuthService
}

func NewAuthHandler(client *ent.Client, authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		client:      client,
		authService: authService,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid request body",
		})
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
	exists, err := h.client.User.Query().
		Where(user.Username(req.Username)).
		Exist(context.Background())
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

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to process password",
		})
	}

	// Create user
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	newUser, err := h.client.User.Create().
		SetUsername(req.Username).
		SetPassword(hashedPassword).
		SetDisplayName(displayName).
		Save(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to create user",
		})
	}

	// Generate token
	token, err := h.authService.CreateToken(newUser.ID, newUser.Username)
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
	var req model.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid request body",
		})
	}

	// Find user
	u, err := h.client.User.Query().
		Where(user.Username(req.Username)).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
				Error: "invalid credentials",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to find user",
		})
	}

	// Verify password
	if err := h.authService.VerifyPassword(u.Password, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
			Error: "invalid credentials",
		})
	}

	// Generate token
	token, err := h.authService.CreateToken(u.ID, u.Username)
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

	u, err := h.client.User.Get(context.Background(), userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get user",
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
