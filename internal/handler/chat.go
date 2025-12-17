package handler

import (
	"context"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/model"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/chat"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/chatmember"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/user"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type ChatHandler struct {
	client *ent.Client
}

func NewChatHandler(client *ent.Client) *ChatHandler {
	return &ChatHandler{client: client}
}

func (h *ChatHandler) CreateChat(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var req model.CreateChatRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid request body",
		})
	}

	if len(req.MemberIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "at least one member is required",
		})
	}

	// Create chat
	newChat, err := h.client.Chat.Create().
		SetName(req.Name).
		SetIsGroup(req.IsGroup).
		SetCreatorID(userID).
		Save(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to create chat",
		})
	}

	// Add creator as admin member
	_, err = h.client.ChatMember.Create().
		SetUserID(userID).
		SetChatID(newChat.ID).
		SetIsAdmin(true).
		Save(context.Background())
	if err != nil {
		// Rollback chat creation
		h.client.Chat.DeleteOneID(newChat.ID).Exec(context.Background())
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to add creator to chat",
		})
	}

	// Add other members
	for _, memberID := range req.MemberIDs {
		if memberID == userID {
			continue // Skip creator as already added
		}
		_, err := h.client.ChatMember.Create().
			SetUserID(memberID).
			SetChatID(newChat.ID).
			SetIsAdmin(false).
			Save(context.Background())
		if err != nil {
			// Continue with other members even if one fails
			continue
		}
	}

	return c.Status(fiber.StatusCreated).JSON(model.ChatResponse{
		ID:        newChat.ID,
		Name:      newChat.Name,
		IsGroup:   newChat.IsGroup,
		CreatorID: userID,
		CreatedAt: newChat.CreatedAt,
		UpdatedAt: newChat.UpdatedAt,
	})
}

func (h *ChatHandler) GetChat(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	chatID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid chat id",
		})
	}

	// Check if user is a member of the chat
	isMember, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(chatID)),
			chatmember.HasUserWith(user.ID(userID)),
		).
		Exist(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to check membership",
		})
	}
	if !isMember {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "you are not a member of this chat",
		})
	}

	// Get chat with members
	chatEntity, err := h.client.Chat.Query().
		Where(chat.ID(chatID)).
		WithCreator().
		WithMembers(func(q *ent.ChatMemberQuery) {
			q.WithUser()
		}).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "chat not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get chat",
		})
	}

	// Build response
	members := make([]model.ChatMemberResponse, 0)
	for _, member := range chatEntity.Edges.Members {
		if member.Edges.User != nil {
			members = append(members, model.ChatMemberResponse{
				UserID:   member.Edges.User.ID,
				Username: member.Edges.User.Username,
				IsAdmin:  member.IsAdmin,
				JoinedAt: member.JoinedAt,
			})
		}
	}

	creatorID := 0
	if chatEntity.Edges.Creator != nil {
		creatorID = chatEntity.Edges.Creator.ID
	}

	return c.JSON(model.ChatDetailResponse{
		ChatResponse: model.ChatResponse{
			ID:        chatEntity.ID,
			Name:      chatEntity.Name,
			IsGroup:   chatEntity.IsGroup,
			CreatorID: creatorID,
			CreatedAt: chatEntity.CreatedAt,
			UpdatedAt: chatEntity.UpdatedAt,
		},
		Members: members,
	})
}

func (h *ChatHandler) ListChats(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	// Get all chat memberships for the user
	memberships, err := h.client.ChatMember.Query().
		Where(chatmember.HasUserWith(user.ID(userID))).
		WithChat(func(q *ent.ChatQuery) {
			q.WithCreator()
		}).
		All(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to list chats",
		})
	}

	chats := make([]model.ChatResponse, 0)
	for _, membership := range memberships {
		if membership.Edges.Chat != nil {
			chatEntity := membership.Edges.Chat
			creatorID := 0
			if chatEntity.Edges.Creator != nil {
				creatorID = chatEntity.Edges.Creator.ID
			}
			chats = append(chats, model.ChatResponse{
				ID:        chatEntity.ID,
				Name:      chatEntity.Name,
				IsGroup:   chatEntity.IsGroup,
				CreatorID: creatorID,
				CreatedAt: chatEntity.CreatedAt,
				UpdatedAt: chatEntity.UpdatedAt,
			})
		}
	}

	return c.JSON(chats)
}

func (h *ChatHandler) UpdateChat(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	chatID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid chat id",
		})
	}

	// Check if user is admin of the chat
	member, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(chatID)),
			chatmember.HasUserWith(user.ID(userID)),
		).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Error: "you are not a member of this chat",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to check membership",
		})
	}
	if !member.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "only admins can update the chat",
		})
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid request body",
		})
	}

	chatEntity, err := h.client.Chat.UpdateOneID(chatID).
		SetName(req.Name).
		Save(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "chat not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to update chat",
		})
	}

	// Get creator ID
	creator, _ := chatEntity.QueryCreator().Only(context.Background())
	creatorID := 0
	if creator != nil {
		creatorID = creator.ID
	}

	return c.JSON(model.ChatResponse{
		ID:        chatEntity.ID,
		Name:      chatEntity.Name,
		IsGroup:   chatEntity.IsGroup,
		CreatorID: creatorID,
		CreatedAt: chatEntity.CreatedAt,
		UpdatedAt: chatEntity.UpdatedAt,
	})
}

func (h *ChatHandler) DeleteChat(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	chatID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid chat id",
		})
	}

	// Check if user is the creator of the chat
	chatEntity, err := h.client.Chat.Get(context.Background(), chatID)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "chat not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get chat",
		})
	}

	creator, _ := chatEntity.QueryCreator().Only(context.Background())
	if creator == nil || creator.ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "only the creator can delete the chat",
		})
	}

	err = h.client.Chat.DeleteOneID(chatID).Exec(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to delete chat",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ChatHandler) AddMembers(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	chatID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid chat id",
		})
	}

	// Check if user is admin of the chat
	member, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(chatID)),
			chatmember.HasUserWith(user.ID(userID)),
		).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Error: "you are not a member of this chat",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to check membership",
		})
	}
	if !member.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "only admins can add members",
		})
	}

	var req model.AddMembersRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid request body",
		})
	}

	// Add members
	for _, memberID := range req.MemberIDs {
		// Check if already a member
		exists, _ := h.client.ChatMember.Query().
			Where(
				chatmember.HasChatWith(chat.ID(chatID)),
				chatmember.HasUserWith(user.ID(memberID)),
			).
			Exist(context.Background())
		if exists {
			continue
		}

		_, err := h.client.ChatMember.Create().
			SetUserID(memberID).
			SetChatID(chatID).
			SetIsAdmin(false).
			Save(context.Background())
		if err != nil {
			continue // Skip if user doesn't exist or other error
		}
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ChatHandler) RemoveMember(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	chatID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid chat id",
		})
	}

	memberID, err := utils.ParamsInt(c, "memberId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid member id",
		})
	}

	// Check if user is admin of the chat
	currentMember, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(chatID)),
			chatmember.HasUserWith(user.ID(userID)),
		).
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
				Error: "you are not a member of this chat",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to check membership",
		})
	}
	if !currentMember.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "only admins can remove members",
		})
	}

	// Cannot remove the creator
	chatEntity, err := h.client.Chat.Get(context.Background(), chatID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get chat",
		})
	}
	creator, _ := chatEntity.QueryCreator().Only(context.Background())
	if creator != nil && creator.ID == memberID {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "cannot remove the creator",
		})
	}

	// Remove member
	_, err = h.client.ChatMember.Delete().
		Where(
			chatmember.HasChatWith(chat.ID(chatID)),
			chatmember.HasUserWith(user.ID(memberID)),
		).
		Exec(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to remove member",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
