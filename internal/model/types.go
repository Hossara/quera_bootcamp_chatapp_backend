package model

import "time"

// Auth models
type LoginRequest struct {
	Username string `json:"username" form:"username" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

type RegisterRequest struct {
	Username    string `json:"username" form:"username" validate:"required,min=3,max=50"`
	Password    string `json:"password" form:"password" validate:"required,min=6"`
	DisplayName string `json:"display_name,omitempty" form:"display_name"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// User models
type UserProfile struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"display_name,omitempty" form:"display_name"`
	Password    string `json:"password,omitempty" form:"password"`
}

// Chat models
type CreateChatRequest struct {
	Name      string `json:"name" form:"name" validate:"required,max=100"`
	IsGroup   bool   `json:"is_group" form:"is_group"`
	MemberIDs []int  `json:"member_ids" form:"member_ids" validate:"required,min=1"`
}

type ChatResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsGroup   bool      `json:"is_group"`
	CreatorID int       `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatDetailResponse struct {
	ChatResponse
	Members []ChatMemberResponse `json:"members"`
}

type ChatMemberResponse struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	IsAdmin  bool      `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
}

type AddMembersRequest struct {
	MemberIDs []int `json:"member_ids" form:"member_ids" validate:"required,min=1"`
}

// Message models
type SendMessageRequest struct {
	Content string `json:"content" form:"content" validate:"required"`
	ChatID  int    `json:"chat_id" form:"chat_id" validate:"required"`
}

type MessageResponse struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	SenderID  int       `json:"sender_id"`
	ChatID    int       `json:"chat_id"`
	IsEdited  bool      `json:"is_edited"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" form:"content" validate:"required"`
}

// WebSocket models
type WSMessage struct {
	Type    string      `json:"type"` // "message", "typing", "read", etc.
	Payload interface{} `json:"payload"`
}

type WSChatMessage struct {
	MessageID int       `json:"message_id"`
	Content   string    `json:"content"`
	SenderID  int       `json:"sender_id"`
	Username  string    `json:"username"`
	ChatID    int       `json:"chat_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
