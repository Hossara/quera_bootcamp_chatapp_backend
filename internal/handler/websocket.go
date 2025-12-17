package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/auth"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/model"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/chat"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/chatmember"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent/user"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

type WebSocketHandler struct {
	client      *ent.Client
	authService *auth.AuthService
	clients     map[int]*websocket.Conn // userID -> connection
	clientsMu   sync.RWMutex
	chatRooms   map[int]map[int]bool // chatID -> map[userID]bool
	roomsMu     sync.RWMutex
	upgrader    websocket.FastHTTPUpgrader
}

func NewWebSocketHandler(client *ent.Client, authService *auth.AuthService) *WebSocketHandler {
	return &WebSocketHandler{
		client:      client,
		authService: authService,
		clients:     make(map[int]*websocket.Conn),
		chatRooms:   make(map[int]map[int]bool),
		upgrader: websocket.FastHTTPUpgrader{
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get token from query parameter
		token := c.Query("token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing token",
			})
		}

		// Verify token
		payload, err := h.authService.VerifyToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		userID := payload.UserID
		username := payload.Username

		// Upgrade to websocket
		if err := h.upgrader.Upgrade(c.RequestCtx(), func(conn *websocket.Conn) {
			defer func() {
				h.clientsMu.Lock()
				delete(h.clients, userID)
				h.clientsMu.Unlock()

				// Remove from all chat rooms
				h.roomsMu.Lock()
				for chatID, members := range h.chatRooms {
					delete(members, userID)
					if len(members) == 0 {
						delete(h.chatRooms, chatID)
					}
				}
				h.roomsMu.Unlock()

				conn.Close()
				log.Printf("User %s (ID: %d) disconnected", username, userID)
			}()

			log.Printf("User %s (ID: %d) connected via WebSocket", username, userID)

			// Register client
			h.clientsMu.Lock()
			h.clients[userID] = conn
			h.clientsMu.Unlock()

			// Handle incoming messages
			for {
				var wsMsg model.WSMessage
				err := conn.ReadJSON(&wsMsg)
				if err != nil {
					log.Printf("Read error for user %d: %v", userID, err)
					break
				}

				switch wsMsg.Type {
				case "message":
					h.handleChatMessage(userID, username, wsMsg.Payload)
				case "join_chat":
					h.handleJoinChat(userID, wsMsg.Payload)
				case "leave_chat":
					h.handleLeaveChat(userID, wsMsg.Payload)
				default:
					log.Printf("Unknown message type: %s", wsMsg.Type)
				}
			}
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "websocket upgrade failed",
			})
		}

		return nil
	}
}

func (h *WebSocketHandler) handleChatMessage(userID int, username string, payload interface{}) {
	// Parse payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return
	}

	var msgReq struct {
		ChatID  int    `json:"chat_id"`
		Content string `json:"content"`
	}
	err = json.Unmarshal(payloadBytes, &msgReq)
	if err != nil {
		log.Printf("Error unmarshaling message request: %v", err)
		return
	}

	// Check if user is a member of the chat
	isMember, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(msgReq.ChatID)),
			chatmember.HasUserWith(user.ID(userID)),
		).
		Exist(context.Background())
	if err != nil || !isMember {
		log.Printf("User %d is not a member of chat %d", userID, msgReq.ChatID)
		return
	}

	// Create message in database
	msg, err := h.client.Message.Create().
		SetContent(msgReq.Content).
		SetSenderID(userID).
		SetChatID(msgReq.ChatID).
		Save(context.Background())
	if err != nil {
		log.Printf("Error creating message: %v", err)
		return
	}

	// Broadcast message to all members of the chat
	wsChatMsg := model.WSChatMessage{
		MessageID: msg.ID,
		Content:   msg.Content,
		SenderID:  userID,
		Username:  username,
		ChatID:    msgReq.ChatID,
		Timestamp: msg.CreatedAt,
	}

	h.broadcastToChat(msgReq.ChatID, model.WSMessage{
		Type:    "message",
		Payload: wsChatMsg,
	})
}

func (h *WebSocketHandler) handleJoinChat(userID int, payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	var req struct {
		ChatID int `json:"chat_id"`
	}
	err = json.Unmarshal(payloadBytes, &req)
	if err != nil {
		return
	}

	// Verify user is a member of the chat
	isMember, err := h.client.ChatMember.Query().
		Where(
			chatmember.HasChatWith(chat.ID(req.ChatID)),
			chatmember.HasUserWith(user.ID(userID)),
		).
		Exist(context.Background())
	if err != nil || !isMember {
		return
	}

	// Add user to chat room
	h.roomsMu.Lock()
	if h.chatRooms[req.ChatID] == nil {
		h.chatRooms[req.ChatID] = make(map[int]bool)
	}
	h.chatRooms[req.ChatID][userID] = true
	h.roomsMu.Unlock()

	log.Printf("User %d joined chat %d", userID, req.ChatID)
}

func (h *WebSocketHandler) handleLeaveChat(userID int, payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	var req struct {
		ChatID int `json:"chat_id"`
	}
	err = json.Unmarshal(payloadBytes, &req)
	if err != nil {
		return
	}

	// Remove user from chat room
	h.roomsMu.Lock()
	if h.chatRooms[req.ChatID] != nil {
		delete(h.chatRooms[req.ChatID], userID)
		if len(h.chatRooms[req.ChatID]) == 0 {
			delete(h.chatRooms, req.ChatID)
		}
	}
	h.roomsMu.Unlock()

	log.Printf("User %d left chat %d", userID, req.ChatID)
}

func (h *WebSocketHandler) broadcastToChat(chatID int, message model.WSMessage) {
	h.roomsMu.RLock()
	members := h.chatRooms[chatID]
	h.roomsMu.RUnlock()

	if members == nil {
		// If no one is in the room, get all members from database
		memberships, err := h.client.ChatMember.Query().
			Where(chatmember.HasChatWith(chat.ID(chatID))).
			All(context.Background())
		if err != nil {
			log.Printf("Error getting chat members: %v", err)
			return
		}

		for _, membership := range memberships {
			sender, _ := membership.QueryUser().Only(context.Background())
			if sender != nil {
				h.sendToUser(sender.ID, message)
			}
		}
		return
	}

	// Send to connected members
	for memberID := range members {
		h.sendToUser(memberID, message)
	}
}

func (h *WebSocketHandler) sendToUser(userID int, message model.WSMessage) {
	h.clientsMu.RLock()
	conn, exists := h.clients[userID]
	h.clientsMu.RUnlock()

	if !exists {
		return
	}

	err := conn.WriteJSON(message)
	if err != nil {
		log.Printf("Error sending message to user %d: %v", userID, err)
		// Connection might be dead, remove it
		h.clientsMu.Lock()
		delete(h.clients, userID)
		h.clientsMu.Unlock()
	}
}

// Helper function to broadcast a system message
func (h *WebSocketHandler) BroadcastSystemMessage(chatID int, message string) error {
	h.broadcastToChat(chatID, model.WSMessage{
		Type: "system",
		Payload: map[string]interface{}{
			"chat_id": chatID,
			"message": message,
		},
	})
	return nil
}

// Helper to notify users about new chat
func (h *WebSocketHandler) NotifyNewChat(userIDs []int, chat model.ChatResponse) {
	message := model.WSMessage{
		Type:    "new_chat",
		Payload: chat,
	}

	for _, userID := range userIDs {
		h.sendToUser(userID, message)
	}
}

// Health check for websocket service
func (h *WebSocketHandler) GetStats() map[string]interface{} {
	h.clientsMu.RLock()
	connectedUsers := len(h.clients)
	h.clientsMu.RUnlock()

	h.roomsMu.RLock()
	activeRooms := len(h.chatRooms)
	h.roomsMu.RUnlock()

	return map[string]interface{}{
		"connected_users": connectedUsers,
		"active_rooms":    activeRooms,
		"status":          "healthy",
	}
}

func (h *WebSocketHandler) HealthCheck(c fiber.Ctx) error {
	stats := h.GetStats()
	return c.JSON(fiber.Map{
		"websocket": stats,
		"message":   fmt.Sprintf("%d users connected, %d rooms active", stats["connected_users"], stats["active_rooms"]),
	})
}
