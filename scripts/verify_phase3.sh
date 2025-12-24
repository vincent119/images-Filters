#!/bin/bash

# Configuration
SERVER_HOST="http://localhost:8080"
SECRET="test-secret-123456"
IMAGE_URL="https://raw.githubusercontent.com/vincent119/images-Filters/main/docs/images/architecture.png"
ENCODED_IMAGE_URL=$(printf %s "$IMAGE_URL" | jq -sRr @uri)

# Cleanup function to kill server on exit
cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID
    fi
}
trap cleanup EXIT

echo "Starting Phase 3 Verification..."

# 0. Start Server
echo "Starting server..."
# Pass necessary env vars for testing
IMG_SECURITY_ENABLED="true" IMG_SECURITY_ALLOW_UNSAFE="true" IMG_SECURITY_SECURITY_KEY="$SECRET" IMG_SECURITY_ALLOWED_SOURCES="*" go run cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server to be ready
echo "Waiting for server to be ready..."
for i in {1..30}; do
    if curl -s "$SERVER_HOST/healthz" > /dev/null; then
        echo "Server is up!"
        break
    fi
    sleep 1
done

# 1. Unsafe Mode Test
echo "Testing /unsafe/ path..."
# Note: Ensure allow_unsafe is true in config or env
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_HOST/unsafe/rs:300x200/$IMAGE_URL")
if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 301 ] || [ "$HTTP_CODE" -eq 308 ]; then
    echo "✅ Unsafe path accessible (Code: $HTTP_CODE)"
else
    echo "⚠️  Unsafe path check returned $HTTP_CODE"
fi

# 2. HMAC Signature Test
echo "Testing HMAC Signature..."

OPTIONS="rs:100x100"
FILTERS="filters:blur(1)"
PATH_TO_SIGN="$OPTIONS/$FILTERS/$IMAGE_URL"

# Build CLI tool if not present (assuming cmd/signer exists from Task 3.2)
# If cmd/signer doesn't exist, we might need to skip or quickly implement a one-off signer here.
# For now assuming it exists or using a direct calculation if possible.
# Wait, Task 3.2 said "建立 CLI 簽名工具", so `cmd/signer` should essentially exist or similar.
# Let's try to build it.
# Always build signer to ensure latest version
echo "Building signer tool..."
go build -o bin/signer ./cmd/signer/main.go || echo "⚠️  Failed to build signer (check path)"

if [ -f "./bin/signer" ]; then
    SIGNED_PATH=$(./bin/signer sign -key "$SECRET" -path "$PATH_TO_SIGN" -quiet)
    echo "Generated Signed Path: $SIGNED_PATH"

    # Request with valid signature
    URL="$SERVER_HOST$SIGNED_PATH"
    echo "Requesting: $URL"
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$URL")
    if [ "$HTTP_CODE" -eq 200 ]; then
        echo "✅ Signed URL verification passed"
    else
        echo "❌ Signed URL verification failed (Code: $HTTP_CODE). Check HMAC secret."
    fi
else
    echo "⚠️  Skipping signature verification test (signer tool not found)"
fi

# Request with invalid signature
INVALID_SIG="invalid-signature-123"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_HOST/$INVALID_SIG/$PATH_TO_SIGN")
if [ "$HTTP_CODE" -eq 403 ]; then
    echo "✅ Invalid signature rejected (403)"
else
    echo "❌ Invalid signature check failed (Code: $HTTP_CODE), expected 403"
fi

echo "Phase 3 Verification Script Completed."

# 3. Source Whitelist Test
echo "Testing Source Whitelist..."
# Restart server with allowed sources
if [ ! -z "$SERVER_PID" ]; then
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null
fi

echo "Restarting server with restricted sources..."
IMG_SECURITY_ENABLED="true" IMG_SECURITY_ALLOW_UNSAFE="true" IMG_SECURITY_SECURITY_KEY="$SECRET" IMG_SECURITY_ALLOWED_SOURCES="raw.githubusercontent.com" go run cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server
for i in {1..30}; do
    if curl -s "$SERVER_HOST/healthz" > /dev/null; then
        break
    fi
    sleep 1
done

# Valid Source
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_HOST/unsafe/rs:300x200/$IMAGE_URL")
if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 301 ] || [ "$HTTP_CODE" -eq 308 ]; then
    echo "✅ Allowed source access passed (Code: $HTTP_CODE)"
else
    echo "❌ Allowed source access failed (Code: $HTTP_CODE)"
fi

# Invalid Source
INVALID_SOURCE_URL="https://google.com/images/logo.png"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER_HOST/unsafe/rs:300x200/$INVALID_SOURCE_URL")
if [ "$HTTP_CODE" -eq 403 ]; then
    echo "✅ Blocked source rejected (403)"
else
    echo "❌ Blocked source check failed (Code: $HTTP_CODE), expected 403"
fi

# 4. Storage & Security Unit Tests
echo "Running Unit Tests..."
go test -v ./internal/storage/... ./internal/security/...
if [ $? -eq 0 ]; then
    echo "✅ Unit tests passed"
else
    echo "❌ Unit tests failed"
fi

echo "Phase 3 Full Verification Completed."
