#!/bin/bash

# Simple integration test for DayZ Server Management API

set -e

BASE_URL="http://localhost:8080/api"
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin123"

echo "Testing DayZ Server Management API"

# Test 1: Health check
echo "1. Testing health endpoint..."
curl -s "${BASE_URL}/health" | grep -q "ok" && echo "✓ Health check passed" || echo "✗ Health check failed"

# Test 2: Setup admin user
echo "2. Setting up admin user..."
SETUP_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/setup" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")

TOKEN=$(echo $SETUP_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -n "$TOKEN" ]; then
  echo "✓ Admin user setup successful"
else
  echo "✗ Admin user setup failed"
  echo "Response: $SETUP_RESPONSE"
  exit 1
fi

# Test 3: Login
echo "3. Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")

NEW_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -n "$NEW_TOKEN" ]; then
  echo "✓ Login successful"
  TOKEN=$NEW_TOKEN
else
  echo "✗ Login failed"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

# Test 4: Create server
echo "4. Creating server..."
CREATE_RESPONSE=$(curl -s -X POST "${BASE_URL}/servers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"TestServer","port":2302}')

SERVER_ID=$(echo $CREATE_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)
if [ -n "$SERVER_ID" ]; then
  echo "✓ Server creation successful (ID: $SERVER_ID)"
else
  echo "✗ Server creation failed"
  echo "Response: $CREATE_RESPONSE"
  exit 1
fi

# Test 5: List servers
echo "5. Listing servers..."
LIST_RESPONSE=$(curl -s -X GET "${BASE_URL}/servers" \
  -H "Authorization: Bearer $TOKEN")
if echo $LIST_RESPONSE | grep -q "TestServer"; then
  echo "✓ Server listing successful"
else
  echo "✗ Server listing failed"
  echo "Response: $LIST_RESPONSE"
fi

# Test 6: Get server details
echo "6. Getting server details..."
DETAILS_RESPONSE=$(curl -s -X GET "${BASE_URL}/servers/$SERVER_ID" \
  -H "Authorization: Bearer $TOKEN")
if echo $DETAILS_RESPONSE | grep -q "TestServer"; then
  echo "✓ Server details retrieval successful"
else
  echo "✗ Server details retrieval failed"
  echo "Response: $DETAILS_RESPONSE"
fi

# Test 7: Get server status
echo "7. Getting server status..."
STATUS_RESPONSE=$(curl -s -X GET "${BASE_URL}/servers/$SERVER_ID/status" \
  -H "Authorization: Bearer $TOKEN")
if echo $STATUS_RESPONSE | grep -q "stopped"; then
  echo "✓ Server status retrieval successful"
else
  echo "✗ Server status retrieval failed"
  echo "Response: $STATUS_RESPONSE"
fi

# Test 8: Get FTP credentials
echo "8. Getting FTP credentials..."
FTP_RESPONSE=$(curl -s -X GET "${BASE_URL}/servers/$SERVER_ID/ftp/credentials" \
  -H "Authorization: Bearer $TOKEN")
if echo $FTP_RESPONSE | grep -q "username"; then
  echo "✓ FTP credentials retrieval successful"
else
  echo "✗ FTP credentials retrieval failed"
  echo "Response: $FTP_RESPONSE"
fi

# Test 9: Delete server
echo "9. Deleting server..."
DELETE_RESPONSE=$(curl -s -X DELETE "${BASE_URL}/servers/$SERVER_ID" \
  -H "Authorization: Bearer $TOKEN")
if echo $DELETE_RESPONSE | grep -q "success"; then
  echo "✓ Server deletion successful"
else
  echo "✗ Server deletion failed"
  echo "Response: $DELETE_RESPONSE"
fi

echo "API Integration Tests Completed!"
echo "Note: Some tests may fail if server binaries are not set up"