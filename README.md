# Chat App Backend

A real-time chat application backend built with Go, Fiber v3, WebSocket, and PostgreSQL.

## Features

- **User Authentication**: Username/password login with PASETO tokens
- **User Management**: Full CRUD operations for user profiles
- **Chat Management**: Create group chats and direct messages
- **Real-time Messaging**: WebSocket support for instant messaging
- **Message Management**: Send, edit, delete messages
- **Member Management**: Add/remove members from group chats

## Tech Stack

- **Go 1.25+**: Programming language
- **Fiber v3**: Web framework
- **Ent**: ORM for database operations
- **PostgreSQL**: Database
- **PASETO**: Secure token-based authentication
- **WebSocket**: Real-time communication
- **Docker Compose**: Container orchestration
- **Air**: Hot reload for development
- **Makefile**: Build automation

## Prerequisites

- Go 1.25 or higher
- Docker and Docker Compose
- Make

## Project Structure

```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── auth/            # Authentication service (PASETO)
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP and WebSocket handlers
│   ├── middleware/      # HTTP middleware
│   ├── model/           # Request/response models
│   └── repository/      # Ent ORM schemas and generated code
│       ├── ent/         # Generated Ent code
│       └── schema/      # Ent schema definitions
├── pkg/
│   └── utils/           # Utility functions
├── docker-compose.yml   # PostgreSQL setup
├── Makefile            # Build commands
├── .air.toml           # Air hot reload config
└── .env.example        # Environment variables template
```

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/Hossara/quera_bootcamp_chatapp_backend.git
cd quera_bootcamp_chatapp_backend
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env if needed
```

### 3. Start PostgreSQL with Docker Compose

```bash
make docker-up
```

### 4. Install dependencies

```bash
make deps
```

### 5. Run the application

For development with hot reload:
```bash
make dev
```

Or run directly:
```bash
make run
```

## Available Make Commands

```bash
make help          # Show all available commands
make run           # Run the application
make build         # Build the application
make test          # Run tests
make clean         # Clean build artifacts
make docker-up     # Start docker containers
make docker-down   # Stop docker containers
make docker-logs   # View docker logs
make ent-generate  # Generate Ent code
make ent-new       # Create new Ent schema
make air-init      # Initialize Air configuration
make dev           # Run with Air hot reload
make install-tools # Install required tools
make deps          # Download dependencies
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/me` - Get current user (authenticated)

### Users

- `GET /api/v1/users` - List all users
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `POST /api/v1/users/last-seen` - Update last seen

### Chats

- `POST /api/v1/chats` - Create a new chat
- `GET /api/v1/chats` - List user's chats
- `GET /api/v1/chats/:id` - Get chat details
- `PUT /api/v1/chats/:id` - Update chat
- `DELETE /api/v1/chats/:id` - Delete chat
- `POST /api/v1/chats/:id/members` - Add members to chat
- `DELETE /api/v1/chats/:id/members/:memberId` - Remove member from chat

### Messages

- `POST /api/v1/messages` - Send a message
- `GET /api/v1/messages/:id` - Get message by ID
- `GET /api/v1/messages/chat/:chatId` - List messages in chat
- `PUT /api/v1/messages/:id` - Update message
- `DELETE /api/v1/messages/:id` - Delete message

### WebSocket

- `GET /ws?token=<auth_token>` - WebSocket connection for real-time chat
- `GET /ws/health` - WebSocket health check

## WebSocket Protocol

### Connection

Connect to WebSocket endpoint with authentication token:
```
ws://localhost:3000/ws?token=YOUR_AUTH_TOKEN
```

### Message Types

#### Send Message
```json
{
  "type": "message",
  "payload": {
    "chat_id": 1,
    "content": "Hello, World!"
  }
}
```

#### Join Chat Room
```json
{
  "type": "join_chat",
  "payload": {
    "chat_id": 1
  }
}
```

#### Leave Chat Room
```json
{
  "type": "leave_chat",
  "payload": {
    "chat_id": 1
  }
}
```

### Receiving Messages

```json
{
  "type": "message",
  "payload": {
    "message_id": 123,
    "content": "Hello, World!",
    "sender_id": 1,
    "username": "john_doe",
    "chat_id": 1,
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

## Authentication

All protected endpoints require an `Authorization` header:

```
Authorization: Bearer <your_token>
```

Tokens are generated using PASETO v2 and expire after 24 hours (configurable).

## Database Schema

### User
- `id`: Primary key
- `username`: Unique username
- `password`: Hashed password
- `display_name`: Display name
- `created_at`: Creation timestamp
- `updated_at`: Update timestamp
- `last_seen`: Last activity timestamp

### Chat
- `id`: Primary key
- `name`: Chat name
- `is_group`: Group or direct message
- `creator_id`: User who created the chat
- `created_at`: Creation timestamp
- `updated_at`: Update timestamp

### Message
- `id`: Primary key
- `content`: Message content
- `sender_id`: User who sent the message
- `chat_id`: Chat the message belongs to
- `is_edited`: Whether message was edited
- `created_at`: Creation timestamp
- `updated_at`: Update timestamp

### ChatMember
- `user_id`: User ID
- `chat_id`: Chat ID
- `is_admin`: Admin status
- `joined_at`: Join timestamp

## Development

### Hot Reload

The project uses Air for hot reload during development:

```bash
make dev
```

Air will automatically rebuild and restart the server when you modify Go files.

### Adding New Ent Schemas

1. Create a new schema:
```bash
make ent-new name=YourSchema
```

2. Edit the schema file in `internal/repository/schema/yourschema.go`

3. Generate Ent code:
```bash
make ent-generate
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `3000` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `chatapp` |
| `DB_PASSWORD` | Database password | `chatapp123` |
| `DB_NAME` | Database name | `chatapp_db` |
| `DB_SSLMODE` | Database SSL mode | `disable` |
| `PASETO_KEY` | PASETO encryption key (32 bytes) | `12345678901234567890123456789012` |
| `TOKEN_EXPIRATION` | Token expiration in hours | `24` |

## Testing

Run tests:
```bash
make test
```

## Building for Production

Build the binary:
```bash
make build
```

The binary will be created in `bin/server`.

## Docker Support

Start the PostgreSQL database:
```bash
make docker-up
```

Stop the database:
```bash
make docker-down
```

View logs:
```bash
make docker-logs
```

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request