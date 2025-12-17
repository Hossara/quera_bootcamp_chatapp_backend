# Quick Start Guide

Get the chat app backend running in 5 minutes!

## Prerequisites
- Go 1.25+ installed
- Docker installed and running
- Make installed (usually pre-installed on Linux/Mac)

## Steps

### 1. Clone and Setup
```bash
git clone https://github.com/Hossara/quera_bootcamp_chatapp_backend.git
cd quera_bootcamp_chatapp_backend
cp .env.example .env
```

### 2. Start Database
```bash
make docker-up
```

Wait a few seconds for PostgreSQL to be ready.

### 3. Install Dependencies
```bash
make deps
```

### 4. Run the Application
```bash
make run
```

You should see:
```
Starting server on 0.0.0.0:3000
```

### 5. Test the API
Open a new terminal and run:
```bash
# Test health
curl http://localhost:3000/health

# Register a user
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123","display_name":"Alice"}'
```

## Full Test Suite
Run the complete test script:
```bash
chmod +x test_api.sh
./test_api.sh
```

## Development Mode
For hot reload during development:
```bash
make dev
```

## Common Commands
```bash
make help          # Show all available commands
make build         # Build binary
make docker-down   # Stop PostgreSQL
make clean         # Clean build artifacts
```

## What's Next?
1. Read the [README.md](README.md) for full documentation
2. Check [IMPLEMENTATION.md](IMPLEMENTATION.md) for architecture details
3. Explore the API endpoints
4. Test WebSocket functionality
5. Deploy to production!

## Troubleshooting

### Port already in use
```bash
# Change port in .env
SERVER_PORT=3001
```

### PostgreSQL connection error
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Restart if needed
make docker-down
make docker-up
```

### Build errors
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

## API Examples

### Register User
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "password123",
    "display_name": "Alice Smith"
  }'
```

### Login
```bash
TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "password123"
  }' | jq -r '.token')
```

### Create Chat
```bash
curl -X POST http://localhost:3000/api/v1/chats \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Chat",
    "is_group": true,
    "member_ids": [1]
  }'
```

### Send Message
```bash
curl -X POST http://localhost:3000/api/v1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "chat_id": 1,
    "content": "Hello, World!"
  }'
```

## WebSocket Connection
```javascript
// JavaScript example
const token = "your-auth-token";
const ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

ws.onopen = () => {
  console.log('Connected!');
  
  // Join a chat
  ws.send(JSON.stringify({
    type: "join_chat",
    payload: { chat_id: 1 }
  }));
  
  // Send a message
  ws.send(JSON.stringify({
    type: "message",
    payload: {
      chat_id: 1,
      content: "Hello from WebSocket!"
    }
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

## Need Help?
- Check [README.md](README.md) for detailed documentation
- See [IMPLEMENTATION.md](IMPLEMENTATION.md) for technical details
- Review API examples in this guide

Happy coding! 🚀
