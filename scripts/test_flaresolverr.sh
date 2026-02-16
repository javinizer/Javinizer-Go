#!/bin/bash

# Test script to verify FlareSolverr is working
# Usage: ./scripts/test_flaresolverr.sh

echo "Testing FlareSolverr connection..."

# Check if FlareSolverr is running
if ! curl -s http://localhost:8191/v1 > /dev/null 2>&1; then
    echo "ERROR: FlareSolverr is not running at http://localhost:8191/v1"
    echo "Start FlareSolverr with:"
    echo "  docker run -p 8191:8191 -e LOG_LEVEL=info ghcr.io/flaresolverr/flaresolverr:latest"
    exit 1
fi

echo "FlareSolverr is running!"

# Test basic request
echo ""
echo "Test 1: Testing basic URL resolution..."
RESPONSE=$(curl -s -X POST http://localhost:8191/v1 \
    -H "Content-Type: application/json" \
    -d '{
        "cmd": "request.get",
        "url": "https://httpbin.org/get",
        "maxTimeout": 10
    }')

if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo "SUCCESS: Basic request returned OK status"
else
    echo "ERROR: Basic request failed"
    echo "Response: $RESPONSE"
fi

# Test with JavLibrary URL (this will be slow due to Cloudflare)
echo ""
echo "Test 2: Testing JavLibrary URL (may take 30-60 seconds)..."
RESPONSE=$(curl -s -X POST http://localhost:8191/v1 \
    -H "Content-Type: application/json" \
    -d '{
        "cmd": "request.get",
        "url": "http://www.javlibrary.com/vl_searchbyid.php?keyword=IPX-123",
        "maxTimeout": 60
    }')

if echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo "SUCCESS: JavLibrary request returned OK status"

    # Check if we got HTML response
    if echo "$RESPONSE" | grep -q "<html"; then
        HTML_LENGTH=$(echo "$RESPONSE" | grep -o '<html' | wc -c)
        echo "Received HTML response ($HTML_LENGTH <html> tags)"

        # Try to extract title if present
        if echo "$RESPONSE" | grep -q -i "<title>"; then
            TITLE=$(echo "$RESPONSE" | grep -oP '<title>.*</title>' | sed 's/<title>//g' | sed 's/<\/title>//g' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
            if [ -n "$TITLE" ]; then
                echo "Title found: $TITLE"
            fi
        fi
    fi
else
    echo "ERROR: JavLibrary request failed"
    echo "Response: $RESPONSE"
fi

echo ""
echo "Test complete!"
