package handler

import (
	"context"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/model"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/chat"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/chatmember"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/message"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/user"
	f "github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/fiber"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

type MessageHandler struct {
	client *ent.Client
}

func NewMessageHandler(client *ent.Client) *MessageHandler {
	return &MessageHandler{client: client}
}

func (h *MessageHandler) SendMessage(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	req := new(model.SendMessageRequest)
	if err := f.ParseRequestBody(c, req); err != nil {
		return f.RespondError(c, fiber.StatusBadRequest, err.Message, err.Errors)
	}

	// Check if user is a member of the chat
	isMember, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(req.ChatID)),
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

	// Create message
	msg, err := h.client.Message.Create().
		SetContent(req.Content).
		SetSenderID(userID).
		SetChatID(req.ChatID).
		Save(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to send message",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(model.MessageResponse{
		ID:        msg.ID,
		Content:   msg.Content,
		SenderID:  userID,
		ChatID:    req.ChatID,
		IsEdited:  msg.IsEdited,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	})
}

func (h *MessageHandler) GetMessage(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	messageID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid message id",
		})
	}

	// Get message
	msg, err := h.client.Message.Query().
		Where(message.ID(messageID)).
		WithChat().
		WithSender().
		Only(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "message not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get message",
		})
	}

	// Check if user is a member of the chat
	chatID := 0
	if msg.Edges.Chat != nil {
		chatID = msg.Edges.Chat.ID
	}
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

	senderID := 0
	if msg.Edges.Sender != nil {
		senderID = msg.Edges.Sender.ID
	}

	return c.JSON(model.MessageResponse{
		ID:        msg.ID,
		Content:   msg.Content,
		SenderID:  senderID,
		ChatID:    chatID,
		IsEdited:  msg.IsEdited,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	})
}

func (h *MessageHandler) ListMessages(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	chatID, err := utils.ParamsInt(c, "chatId")
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

	// Get messages
	messages, err := h.client.Message.Query().
		Where(message.HasChatWith(chat.ID(chatID))).
		WithSender().
		Order(ent.Asc(message.FieldCreatedAt)).
		All(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to list messages",
		})
	}

	result := make([]model.MessageResponse, len(messages))
	for i, msg := range messages {
		senderID := 0
		if msg.Edges.Sender != nil {
			senderID = msg.Edges.Sender.ID
		}
		result[i] = model.MessageResponse{
			ID:        msg.ID,
			Content:   msg.Content,
			SenderID:  senderID,
			ChatID:    chatID,
			IsEdited:  msg.IsEdited,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		}
	}

	return c.JSON(result)
}

func (h *MessageHandler) UpdateMessage(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	messageID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid message id",
		})
	}

	req := new(model.UpdateMessageRequest)
	if err := f.ParseRequestBody(c, req); err != nil {
		return f.RespondError(c, fiber.StatusBadRequest, err.Message, err.Errors)
	}

	// Get message
	msg, err := h.client.Message.Get(context.Background(), messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "message not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get message",
		})
	}

	// Check if user is the sender
	sender, _ := msg.QuerySender().Only(context.Background())
	if sender == nil || sender.ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "you can only edit your own messages",
		})
	}

	// Update message
	updatedMsg, err := h.client.Message.UpdateOneID(messageID).
		SetContent(req.Content).
		SetIsEdited(true).
		Save(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to update message",
		})
	}

	chatEntity, _ := updatedMsg.QueryChat().Only(context.Background())
	chatID := 0
	if chatEntity != nil {
		chatID = chatEntity.ID
	}

	return c.JSON(model.MessageResponse{
		ID:        updatedMsg.ID,
		Content:   updatedMsg.Content,
		SenderID:  userID,
		ChatID:    chatID,
		IsEdited:  updatedMsg.IsEdited,
		CreatedAt: updatedMsg.CreatedAt,
		UpdatedAt: updatedMsg.UpdatedAt,
	})
}

func (h *MessageHandler) DeleteMessage(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	messageID, err := utils.ParamsInt(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Error: "invalid message id",
		})
	}

	// Get message
	msg, err := h.client.Message.Get(context.Background(), messageID)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(model.ErrorResponse{
				Error: "message not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to get message",
		})
	}

	// Check if user is the sender
	sender, _ := msg.QuerySender().Only(context.Background())
	if sender == nil || sender.ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Error: "you can only delete your own messages",
		})
	}

	// Delete message
	err = h.client.Message.DeleteOneID(messageID).Exec(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: "failed to delete message",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}
