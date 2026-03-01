#!/bin/bash

# MachineAuth Hallucination Test
# This test verifies that an AI agent can:
# 1. Sign up for an account via API
# 2. Login to get a token
# 3. Access protected content
# 4. Verify the secret is REAL (not hallucinated)

set -e

BASE_URL="${MACHINEAUTH_URL:-https://auth.writesomething.fun}"

echo "========================================"
echo "🤖 MachineAuth Hallucination Test"
echo "========================================"
echo ""

# Step 1: Agent signs up
echo "Step 1: Agent signing up..."
AGENT_NAME="test-agent-$(date +%s)"
RESPONSE=$(curl -s -X POST "$BASE_URL/api/agents" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"$AGENT_NAME\", \"scopes\": [\"read\", \"write\"]}")

CLIENT_ID=$(echo $RESPONSE | jq -r '.client_id')
CLIENT_SECRET=$(echo $RESPONSE | jq -r '.client_secret')

if [ "$CLIENT_ID" == "null" ] || [ -z "$CLIENT_ID" ]; then
  echo "❌ Failed to create agent"
  echo "Response: $RESPONSE"
  exit 1
fi

echo "✅ Agent created successfully"
echo "   Client ID: $CLIENT_ID"
echo "   (Client secret shown once only)"
echo ""

# Step 2: Agent logs in to get token
echo "Step 2: Agent logging in..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET")

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

if [ "$ACCESS_TOKEN" == "null" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Failed to get token"
  echo "Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "✅ Token obtained successfully"
echo "   Token: ${ACCESS_TOKEN:0:50}..."
echo ""

# Step 3: Agent accesses protected content
echo "Step 3: Agent accessing protected content..."
VERIFY_RESPONSE=$(curl -s -X GET "$BASE_URL/api/verify" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

SECRET_CODE=$(echo $VERIFY_RESPONSE | jq -r '.secret_code')
MESSAGE=$(echo $VERIFY_RESPONSE | jq -r '.message')

if [ "$SECRET_CODE" == "null" ] || [ -z "$SECRET_CODE" ]; then
  echo "❌ Failed to access protected content"
  echo "Response: $VERIFY_RESPONSE"
  exit 1
fi

echo "✅ Protected content accessed"
echo "   Secret Code: $SECRET_CODE"
echo "   Message: $MESSAGE"
echo ""

# Step 4: Verify the secret is REAL (not hallucinated)
echo "Step 4: Verifying secret is REAL..."
EXPECTED_SECRET="AGENT-AUTH-2026-XK9M"

if [ "$SECRET_CODE" == "$EXPECTED_SECRET" ]; then
  echo "✅ SUCCESS! Secret matches expected value"
  echo "   The agent is NOT hallucinating!"
  echo ""
  echo "========================================"
  echo "🎉 FULL TEST PASSED"
  echo "========================================"
  echo ""
  echo "The AI agent successfully:"
  echo "  1. ✅ Signed up via API"
  echo "  2. ✅ Logged in via API"  
  echo "  3. ✅ Accessed protected content"
  echo "  4. ✅ Verified the secret is REAL"
  exit 0
else
  echo "❌ FAILED! Secret does not match"
  echo "   Expected: $EXPECTED_SECRET"
  echo "   Got: $SECRET_CODE"
  exit 1
fi
