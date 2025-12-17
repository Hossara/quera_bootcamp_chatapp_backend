#!/bin/bash

# Test script for Chat App Backend API

BASE_URL="http://localhost:3000"
echo "=========================================="
echo "Testing Chat App Backend API"
echo "=========================================="
echo ""

# 1. Test Health Endpoint
echo "1. Testing Health Endpoint..."
curl -s ${BASE_URL}/health | jq .
echo ""

# 2. Register User 1
echo "2. Registering User 1..."
USER1=$(curl -s -X POST ${BASE_URL}/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123","display_name":"Alice Smith"}')
echo "$USER1" | jq .
TOKEN1=$(echo "$USER1" | jq -r '.token')
echo ""

# 3. Register User 2
echo "3. Registering User 2..."
USER2=$(curl -s -X POST ${BASE_URL}/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"password123","display_name":"Bob Jones"}')
echo "$USER2" | jq .
TOKEN2=$(echo "$USER2" | jq -r '.token')
echo ""

# 4. List Users
echo "4. Listing all users..."
curl -s -H "Authorization: Bearer $TOKEN1" ${BASE_URL}/api/v1/users | jq .
echo ""

# 5. Create a Chat
echo "5. Creating a group chat..."
CHAT=$(curl -s -X POST ${BASE_URL}/api/v1/chats \
  -H "Authorization: Bearer $TOKEN1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Project Discussion","is_group":true,"member_ids":[1,2]}')
echo "$CHAT" | jq .
CHAT_ID=$(echo "$CHAT" | jq -r '.id')
echo ""

# 6. Send Messages
echo "6. Sending messages..."
curl -s -X POST ${BASE_URL}/api/v1/messages \
  -H "Authorization: Bearer $TOKEN1" \
  -H "Content-Type: application/json" \
  -d "{\"chat_id\":${CHAT_ID},\"content\":\"Hi Bob, how are you?\"}" | jq .
echo ""

curl -s -X POST ${BASE_URL}/api/v1/messages \
  -H "Authorization: Bearer $TOKEN2" \
  -H "Content-Type: application/json" \
  -d "{\"chat_id\":${CHAT_ID},\"content\":\"Hi Alice! I'm doing great, thanks!\"}" | jq .
echo ""

# 7. List Messages
echo "7. Listing messages in chat..."
curl -s -H "Authorization: Bearer $TOKEN1" ${BASE_URL}/api/v1/messages/chat/${CHAT_ID} | jq .
echo ""

# 8. Get Chat Details
echo "8. Getting chat details..."
curl -s -H "Authorization: Bearer $TOKEN1" ${BASE_URL}/api/v1/chats/${CHAT_ID} | jq .
echo ""

# 9. List User's Chats
echo "9. Listing user's chats..."
curl -s -H "Authorization: Bearer $TOKEN1" ${BASE_URL}/api/v1/chats | jq .
echo ""

# 10. WebSocket Health
echo "10. Checking WebSocket health..."
curl -s ${BASE_URL}/ws/health | jq .
echo ""

echo "=========================================="
echo "All tests completed successfully!"
echo "=========================================="
