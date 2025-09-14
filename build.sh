#!/bin/bash

echo "🔧 Setting up Incident Commander Game..."

# Create directories
mkdir -p web/static web/images

# Build WebAssembly
echo "🏗️  Building WebAssembly module..."
cd /Users/nathan.nam/Documents/GitHub/NathanNam/incident-commander-game-no-instrumentation
GOOS=js GOARCH=wasm go build -o web/static/game.wasm cmd/game/main.go

# Copy WebAssembly support
echo "📋 Copying WebAssembly support files..."
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/static/

echo "✅ Build complete!"
echo "🚀 Run 'go run cmd/server/main.go' to start the server"