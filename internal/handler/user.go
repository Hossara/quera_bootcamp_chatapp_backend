package handler

import (
	"context"
	"time"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/auth"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/model"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/user"
	f "github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/fiber"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	client      *ent.Client
	authService *auth.Service
}

func NewUserHandler(client *ent.Client, authService *auth.Service) *UserHandler {
	return &UserHandler{
		client:      client,
		authService: authService,
	}
}

func (h *UserHandler) GetUser(c fiber.Ctx) error {
	id, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid user id",
		})
	}

	u, err := h.client.User.Get(context.Background(), id)
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

func (h *UserHandler) ListUsers(c fiber.Ctx) error {
	users, err := h.client.User.Query().
		Order(ent.Asc(user.FieldUsername)).
		All(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to list users",
		})
	}

	profiles := make([]model.UserProfile, len(users))
	for i, u := range users {
		profiles[i] = model.UserProfile{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			CreatedAt:   u.CreatedAt,
			LastSeen:    u.LastSeen,
		}
	}

	return c.JSON(profiles)
}

func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	id, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid user id",
		})
	}

	// Users can only update their own profile
	if userID != id {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "you can only update your own profile",
		})
	}

	req := new(model.UpdateUserRequest)
	if err := f.ParseRequestBody(c, req); err != nil {
		return f.RespondError(c, fiber.StatusBadRequest, err.Message, err.Errors)
	}

	update := h.client.User.UpdateOneID(id)

	if req.DisplayName != "" {
		update.SetDisplayName(req.DisplayName)
	}

	if req.Password != "" {
		if len(req.Password) < 6 {
			return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
				Error: "password must be at least 6 characters",
			})
		}
		hashedPassword, err := h.authService.HashPassword(req.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
				Error: "failed to process password",
			})
		}
		update.SetPassword(hashedPassword)
	}

	u, err := update.Save(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to update user",
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

func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	id, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid user id",
		})
	}

	// Users can only delete their own profile
	if userID != id {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "you can only delete your own profile",
		})
	}

	err = h.client.User.DeleteOneID(id).Exec(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to delete user",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *UserHandler) UpdateLastSeen(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	now := time.Now()
	_, err := h.client.User.UpdateOneID(userID).
		SetLastSeen(now).
		Save(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to update last seen",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
