.PHONY: build run clean wasm server setup

# Default target
all: build

# Setup creates necessary directories and copies assets
setup:
	@echo "🔧 Setting up project structure..."
	@mkdir -p web/static web/images
	@echo "📁 Directories created"
	@if [ -f "web/images/o11y_alert.png" ]; then \
		echo "🖼️  o11y_alert.png already exists in web/images/"; \
	elif [ -f "o11y_alert.png" ]; then \
		cp o11y_alert.png web/images/; \
		echo "🖼️  Copied o11y_alert.png to web/images/"; \
	else \
		echo "ℹ️  o11y_alert.png not found - game will use fallback graphics"; \
	fi

# Build WebAssembly and prepare static files
build: setup
	@echo "🏗️  Building WebAssembly module..."
	@GOOS=js GOARCH=wasm go build -o web/static/game.wasm cmd/game/main.go
	@echo "📋 Copying WebAssembly support files..."
	@GOROOT=$$(go env GOROOT); \
	if [ -f "$$GOROOT/misc/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/misc/wasm/wasm_exec.js" web/static/; \
	else \
		echo "⚠️  wasm_exec.js not found at $$GOROOT/misc/wasm/"; \
		echo "🔍 Checking alternative locations..."; \
		if [ -f "/usr/local/go/misc/wasm/wasm_exec.js" ]; then \
			cp "/usr/local/go/misc/wasm/wasm_exec.js" web/static/; \
			echo "✅ Found wasm_exec.js at /usr/local/go/misc/wasm/"; \
		elif [ -f "/opt/homebrew/lib/go/misc/wasm/wasm_exec.js" ]; then \
			cp "/opt/homebrew/lib/go/misc/wasm/wasm_exec.js" web/static/; \
			echo "✅ Found wasm_exec.js at /opt/homebrew/lib/go/misc/wasm/"; \
		else \
			echo "❌ wasm_exec.js not found. Please check your Go installation."; \
			echo "💡 You can download it from: https://github.com/golang/go/raw/master/misc/wasm/wasm_exec.js"; \
			exit 1; \
		fi; \
	fi
	@echo "✅ Build complete!"

# Build and run the server
run: build
	@echo "🚀 Starting Incident Commander Game Server..."
	@go run cmd/server/main.go

# Run only the server (assumes WebAssembly is already built)
server:
	@echo "🚀 Starting server..."
	@go run cmd/server/main.go

# Build only the WebAssembly module
wasm:
	@echo "🔨 Building WebAssembly module..."
	@mkdir -p web/static
	@GOOS=js GOARCH=wasm go build -o web/static/game.wasm cmd/game/main.go
	@GOROOT=$$(go env GOROOT); \
	if [ -f "$$GOROOT/misc/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/misc/wasm/wasm_exec.js" web/static/; \
	elif [ -f "/usr/local/go/misc/wasm/wasm_exec.js" ]; then \
		cp "/usr/local/go/misc/wasm/wasm_exec.js" web/static/; \
	elif [ -f "/opt/homebrew/lib/go/misc/wasm/wasm_exec.js" ]; then \
		cp "/opt/homebrew/lib/go/misc/wasm/wasm_exec.js" web/static/; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf web/static/game.wasm web/static/wasm_exec.js
	@echo "✅ Clean complete!"

# Development mode - rebuild and restart on changes
dev:
	@echo "👨‍💻 Starting development mode..."
	@echo "Press Ctrl+C to stop"
	@while true; do \
		make run; \
		sleep 2; \
	done

# Test the health endpoint
test-health:
	@echo "🔍 Testing health endpoint..."
	@curl -s http://localhost:8080/health | python -m json.tool || echo "Server not running or health endpoint unavailable"

# Display build information
info:
	@echo "📊 Build Information:"
	@echo "Go Version: $$(go version)"
	@echo "GOOS=js GOARCH=wasm"
	@echo "Project: Incident Commander Game"
	@echo "Port: 8080"
	@echo ""
	@echo "🎯 Development targets:"
	@echo "  make setup       - Create directories and copy assets"
	@echo "  make build       - Build WebAssembly and prepare files"
	@echo "  make run         - Build and run the game server"
	@echo "  make server      - Run server only (no rebuild)"
	@echo "  make wasm        - Build WebAssembly module only"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make dev         - Development mode with auto-restart"
	@echo "  make test-health - Test the health endpoint"
	@echo ""
	@echo "🚀 Ubuntu/EC2 deployment targets:"
	@echo "  make ubuntu-deps    - Install Go and dependencies on Ubuntu"
	@echo "  make build-prod     - Build optimized production version"
	@echo "  make build-ubuntu   - Build Linux binary for Ubuntu"
	@echo "  make run-ubuntu     - Run server on Ubuntu (foreground)"
	@echo "  make run-daemon     - Run server in background"
	@echo "  make stop-daemon    - Stop background server"
	@echo "  make status         - Check daemon server status"
	@echo "  make install-service- Install as systemd service"
	@echo "  make service-start  - Start systemd service"
	@echo "  make service-stop   - Stop systemd service"
	@echo "  make service-status - Check service status"
	@echo "  make service-logs   - View service logs"
	@echo "  make setup-firewall - Configure Ubuntu firewall"
	@echo "  make deploy-ubuntu  - Complete Ubuntu deployment"
	@echo "  make clean-deploy   - Clean deployment files"
	@echo ""
	@echo "  make info           - Show this information"

# Quick development setup
quick: clean build
	@echo "⚡ Quick build complete! Run 'make run' to start the server."

# ========================================
# Ubuntu/EC2 Deployment Targets
# ========================================

# Install dependencies on Ubuntu
ubuntu-deps:
	@echo "🔧 Installing dependencies on Ubuntu..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "📦 Installing Go..."; \
		sudo apt update; \
		sudo apt install -y wget; \
		wget -q https://golang.org/dl/go1.21.3.linux-amd64.tar.gz; \
		sudo rm -rf /usr/local/go; \
		sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz; \
		rm go1.21.3.linux-amd64.tar.gz; \
		echo 'export PATH=$$PATH:/usr/local/go/bin' >> ~/.bashrc; \
		echo "✅ Go installed. Please run 'source ~/.bashrc' or logout/login"; \
	else \
		echo "✅ Go already installed: $$(go version)"; \
	fi
	@if ! command -v git >/dev/null 2>&1; then \
		echo "📦 Installing Git..."; \
		sudo apt install -y git; \
		echo "✅ Git installed"; \
	else \
		echo "✅ Git already installed"; \
	fi
	@echo "🎉 All dependencies ready!"

# Build for production (optimized)
build-prod: setup
	@echo "🏗️  Building WebAssembly module (production)..."
	@GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/static/game.wasm cmd/game/main.go
	@echo "📋 Copying WebAssembly support files..."
	@GOROOT=$$(go env GOROOT); \
	if [ -f "$$GOROOT/misc/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/misc/wasm/wasm_exec.js" web/static/; \
	elif [ -f "/usr/local/go/misc/wasm/wasm_exec.js" ]; then \
		cp "/usr/local/go/misc/wasm/wasm_exec.js" web/static/; \
	else \
		echo "❌ wasm_exec.js not found"; \
		exit 1; \
	fi
	@echo "✅ Production build complete!"

# Build binary for Ubuntu deployment
build-ubuntu: build-prod
	@echo "🏗️  Building server binary for Linux..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o incident-commander-server cmd/server/main.go
	@echo "✅ Ubuntu binary built: incident-commander-server"

# Run on Ubuntu (production mode)
run-ubuntu: build-ubuntu
	@echo "🚀 Starting Incident Commander Game Server (Ubuntu)..."
	@echo "🌐 Server will be available at http://your-server-ip:8080"
	@echo "🔍 Health check: http://your-server-ip:8080/health"
	@./incident-commander-server

# Run in background (daemon mode)
run-daemon: build-ubuntu
	@echo "🚀 Starting server in background..."
	@nohup ./incident-commander-server > incident-commander.log 2>&1 & echo $$! > incident-commander.pid
	@echo "✅ Server started in background (PID: $$(cat incident-commander.pid))"
	@echo "📋 Log file: incident-commander.log"
	@echo "🔍 Check status: make status"

# Check server status
status:
	@if [ -f "incident-commander.pid" ]; then \
		PID=$$(cat incident-commander.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "✅ Server is running (PID: $$PID)"; \
			echo "📊 Memory usage: $$(ps -p $$PID -o rss= | awk '{print $$1/1024 " MB"}')"; \
		else \
			echo "❌ Server not running (stale PID file)"; \
			rm -f incident-commander.pid; \
		fi; \
	else \
		echo "❌ No PID file found. Server not running in daemon mode."; \
	fi

# Stop daemon server
stop-daemon:
	@if [ -f "incident-commander.pid" ]; then \
		PID=$$(cat incident-commander.pid); \
		if ps -p $$PID > /dev/null 2>&1; then \
			kill $$PID; \
			echo "🛑 Server stopped (PID: $$PID)"; \
		else \
			echo "⚠️  Server was not running"; \
		fi; \
		rm -f incident-commander.pid; \
	else \
		echo "❌ No PID file found"; \
	fi

# Install systemd service (Ubuntu)
install-service: build-ubuntu
	@echo "🔧 Installing systemd service..."
	@sudo tee /etc/systemd/system/incident-commander.service > /dev/null <<EOF \
[Unit]\
Description=Incident Commander Game Server\
After=network.target\
\
[Service]\
Type=simple\
User=$$USER\
WorkingDirectory=$$(pwd)\
ExecStart=$$(pwd)/incident-commander-server\
Restart=always\
RestartSec=10\
\
[Install]\
WantedBy=multi-user.target\
EOF
	@sudo systemctl daemon-reload
	@sudo systemctl enable incident-commander.service
	@echo "✅ Service installed! Use 'make service-start' to start"

# Service management
service-start:
	@sudo systemctl start incident-commander.service
	@echo "✅ Service started"

service-stop:
	@sudo systemctl stop incident-commander.service  
	@echo "🛑 Service stopped"

service-restart:
	@sudo systemctl restart incident-commander.service
	@echo "🔄 Service restarted"

service-status:
	@sudo systemctl status incident-commander.service

service-logs:
	@sudo journalctl -u incident-commander.service -f

# Setup Ubuntu firewall (if needed)
setup-firewall:
	@echo "🔥 Configuring UFW firewall..."
	@sudo ufw allow 8080/tcp
	@sudo ufw allow ssh
	@echo "✅ Firewall configured (port 8080 open)"

# Complete Ubuntu deployment
deploy-ubuntu: ubuntu-deps build-ubuntu install-service setup-firewall service-start
	@echo "🎉 Deployment complete!"
	@echo "🌐 Game available at: http://$$(curl -s ifconfig.me):8080"
	@echo "🔍 Health check: http://$$(curl -s ifconfig.me):8080/health"
	@echo "📊 Service status: make service-status"
	@echo "📋 View logs: make service-logs"

# Clean deployment files
clean-deploy:
	@echo "🧹 Cleaning deployment files..."
	@rm -f incident-commander-server incident-commander.pid incident-commander.log
	@echo "✅ Deployment files cleaned"