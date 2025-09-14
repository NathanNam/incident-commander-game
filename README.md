# 🎮 Incident Commander Game

A Snake-like browser game where you play as the **Incident Commander** (o11y mascot) collecting alert notifications. Built with **Go and WebAssembly** for smooth 60 FPS performance on desktop and mobile browsers.

![Game Screenshot](https://img.shields.io/badge/Platform-Web%20Browser-blue?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![WebAssembly](https://img.shields.io/badge/WebAssembly-654FF0?style=for-the-badge&logo=webassembly&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

## 🎯 Game Overview

Control the **Incident Commander** (adorable o11y mascot) as it moves around collecting red alert bubbles. Like Snake, your trail grows with each alert collected, but avoid hitting walls or your own trail! Progress through 10 increasingly challenging levels with obstacles and faster speeds.

### 🎮 **Game Mechanics**
- **Snake-like Movement** - Commander moves continuously in chosen direction
- **Alert Collection** - Collect red alert bubbles (marked with "!") to score points  
- **Growing Trail** - Your resolved incident trail grows with each alert
- **Collision Avoidance** - Don't hit walls, obstacles, or your own trail
- **Level Progression** - Complete 10 levels with increasing difficulty

## 🚀 Quick Start

### **Local Development**
```bash
# Clone the repository
git clone <repository-url>
cd incident-commander-game-no-instrumentation

# Build and run
make run

# Open browser to http://localhost:8080
```

### **Ubuntu/EC2 Deployment** 
```bash
# One-command deployment
make deploy-ubuntu

# Your game will be available at http://YOUR_SERVER_IP:8080
```

## ✨ Features

### 🎯 **Core Gameplay**
- **10 Progressive Levels** - From slow (500ms) to fast (125ms) movement
- **Dynamic Obstacles** - Static barriers, moving obstacles, maze layouts
- **Smart Scoring** - Base points + combo multipliers + time bonuses
- **Level Transitions** - Trail resets between levels, brief completion pause

### 🖥️ **Cross-Platform Support**
- **Desktop Browsers** - Chrome, Firefox, Safari, Edge
- **Mobile Optimized** - iPhone Chrome, Android browsers
- **Touch Controls** - Swipe gestures + on-screen buttons
- **Responsive Design** - No scrolling required on any device

### ⚡ **Technical Excellence** 
- **WebAssembly Performance** - Go compiled to WASM for 60 FPS gameplay
- **Session Isolation** - Each browser tab gets independent game instance
- **Health Monitoring** - Built-in health check endpoint (`/health`)
- **Production Ready** - Optimized builds, systemd service, daemon mode

### 🎨 **Visual Design**
- **O11y Mascot Sprite** - Authentic observability theme character
- **Fallback Graphics** - Blue circle with eyes if image fails to load
- **Modern UI** - Clean, responsive interface with game state indicators
- **Mobile-First Layout** - 60% game canvas, 25% touch controls, 15% stats

## 🎮 Controls

### **Desktop**
- **Arrow Keys** or **WASD** - Move the Incident Commander
- **Space** or **P** - Pause/Resume game
- **R** - Restart game

### **Mobile**
- **Swipe Gestures** - Change direction (up/down/left/right)
- **Tap Canvas** - Pause/Resume
- **On-Screen Buttons** - Alternative touch controls
- **Immediate Response** - Uses `touchstart` events for lag-free control

## 📊 Level Progression

| Level | Speed | Alerts Needed | Obstacles | Special Features |
|-------|-------|---------------|-----------|------------------|
| 1 | 500ms | 5 | None | Learning level |
| 2 | 435ms | 7 | None | Speed increase |
| 3-4 | 370-305ms | 9-11 | Static barriers | Cross patterns |
| 5-6 | 240-175ms | 13-15 | More obstacles | Complex layouts |
| 7-8 | 110-175ms | 17-19 | Random spawns | Dynamic barriers |
| 9-10 | 125ms | 21-25 | Maze layouts | Maximum challenge |

### **Scoring System**
- **Base Points**: 10 per alert
- **Combo Multiplier**: Consecutive collections (1x, 2x, 3x...)
- **Level Completion Bonus**: 100 × level number
- **Time Bonus**: Up to 60 points for fast completion

## 🛠️ Development

### **Build Commands**
```bash
# Development
make run          # Build and run locally
make build        # Build WebAssembly + assets
make clean        # Clean build artifacts
make dev          # Development mode with auto-restart

# Testing  
make test-health  # Test health endpoint
make info         # Show build information
```

### **Project Structure**
```
incident-commander-game/
├── cmd/
│   ├── server/main.go        # HTTP server with CORS + health endpoint
│   └── game/main.go          # WebAssembly entry point + game loop
├── internal/
│   ├── game/game.go          # Core game logic (10 levels, scoring)
│   ├── renderer/renderer.go  # Canvas rendering + mascot graphics
│   └── input/input.go        # Keyboard + touch input handling
├── web/
│   ├── index.html            # iOS-optimized single-page app
│   ├── images/o11y_alert.png # Game mascot sprite
│   └── static/               # Built WebAssembly files
├── Makefile                  # Build + deployment automation
├── DEPLOYMENT.md             # Ubuntu/EC2 deployment guide
└── README.md                 # This file
```

### **Technology Stack**
- **Backend**: Go 1.21+ with `gorilla/websocket`
- **Frontend**: WebAssembly + HTML5 Canvas + Vanilla JavaScript
- **Build System**: Makefile with cross-platform support
- **Deployment**: Systemd service + daemon mode + health checks

## 🚀 Deployment

### **Local Development**
```bash
make run                    # http://localhost:8080
```

### **Ubuntu/EC2 Production**
```bash
# Complete deployment (installs dependencies, builds, configures service)
make deploy-ubuntu

# Individual steps
make ubuntu-deps           # Install Go + dependencies
make build-ubuntu          # Build Linux binary
make install-service       # Install systemd service
make service-start         # Start service
```

### **Service Management**
```bash
make service-status        # Check service status
make service-logs          # View real-time logs
make service-restart       # Restart service
make status               # Check daemon status (if using daemon mode)
```

### **Deployment Options**
1. **Systemd Service** - Production-ready with auto-restart
2. **Daemon Mode** - Background process with PID tracking
3. **Foreground Mode** - Direct execution for testing

## 🌐 Endpoints

- **`GET /`** - Game interface (HTML + WebAssembly)
- **`GET /health`** - Health check endpoint
- **`GET /static/*`** - WebAssembly files (`game.wasm`, `wasm_exec.js`)
- **`GET /images/*`** - Game assets (`o11y_alert.png`)

### **Health Check Response**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-14T12:34:56Z",
  "service": "incident-commander-game"
}
```

## 📱 Mobile Optimization

### **iOS Chrome Specific Features**
- **Proper viewport handling** - `viewport-fit=cover` for iPhone notch
- **Prevents rubber-band effects** - No unwanted scrolling/bouncing
- **Touch-optimized controls** - Immediate `touchstart` response
- **Safe area support** - Padding for notch and home indicator
- **No-scroll design** - Entire game fits on screen without scrolling

### **Responsive Layout**
- **Score Panel**: 15% viewport height - Game stats and level info
- **Game Canvas**: 60% viewport height - Main game area  
- **Touch Controls**: 25% viewport height - Mobile control buttons
- **Auto-scaling** - Canvas resizes to fit any screen size

## 🔧 Configuration

### **Server Configuration**
- **Port**: 8080 (configurable in server code)
- **CORS**: Enabled for WebAssembly files
- **Static Files**: Served from `web/` directory
- **Health Check**: Available at `/health`

### **Game Configuration**
- **Grid Size**: 20×20 cells (configurable in game code)
- **Frame Rate**: Variable based on level (2-8 FPS)
- **Session Management**: Isolated per browser connection
- **Image Assets**: Fallback graphics if mascot image unavailable

## 🧪 Testing

### **Browser Compatibility**
| Browser | Desktop | Mobile | Status |
|---------|---------|---------|---------|
| Chrome | ✅ | ✅ | Full support |
| Safari | ✅ | ✅ | Full support |
| Firefox | ✅ | ✅ | Full support |
| Edge | ✅ | ✅ | Full support |

### **Mobile Testing**
- **iPhone SE** (375×667px) ✅
- **iPhone 12/13/14** (390×844px) ✅  
- **iPhone Pro Max** (428×926px) ✅
- **Android** (360×640px+) ✅

### **Performance Testing**
```bash
# Load test health endpoint
curl http://localhost:8080/health

# Test WebAssembly loading
curl -I http://localhost:8080/static/game.wasm

# Multiple browser sessions
# Each should get independent game instance
```

## 🐛 Troubleshooting

### **Common Issues**

**Game won't load:**
- Check browser console for WebAssembly errors
- Verify `/static/game.wasm` and `/static/wasm_exec.js` exist
- Ensure CORS headers are working

**Mobile controls not working:**
- Verify touch events aren't blocked
- Check that `touchstart` events are firing
- Ensure canvas has `touch-action: none` CSS

**Level transitions stuck:**
- Check browser console for game state errors
- Verify game update loop is running
- Ensure timer-based transitions work

**Ubuntu deployment issues:**
- Run `make ubuntu-deps` to install Go
- Check `make service-logs` for errors
- Verify port 8080 is open: `sudo ufw allow 8080/tcp`

### **Debug Commands**
```bash
# Check service status
make service-status

# View logs
make service-logs

# Test health endpoint
curl http://localhost:8080/health

# Check if binary works
./incident-commander-server
```

## 📈 Performance

### **Metrics**
- **Load Time**: < 3 seconds on 3G connection
- **Frame Rate**: 60 FPS animation, 2-8 FPS game logic
- **Memory Usage**: < 50MB typical
- **WebAssembly Size**: < 2MB compressed
- **Battery Efficient**: Optimized game loop for mobile

### **Optimizations**
- **Compiled WebAssembly** - Go code runs at near-native speed
- **Efficient Rendering** - Canvas-based graphics with minimal DOM manipulation
- **Session Isolation** - Each game instance runs independently
- **Production Builds** - Optimized binaries with `-ldflags="-s -w"`

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/amazing-feature`
3. **Make changes and test**: `make run`
4. **Test on mobile**: Verify touch controls work
5. **Submit pull request**

### **Development Guidelines**
- Follow Go best practices and formatting
- Test on both desktop and mobile
- Maintain 60 FPS performance
- Keep WebAssembly binary size minimal
- Update documentation for new features

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Go Team** - For excellent WebAssembly support
- **Observability Community** - For the adorable o11y mascot
- **Snake Game** - Classic gameplay inspiration
- **Canvas API** - For smooth graphics rendering

## 📞 Support

- **Issues**: Report bugs and feature requests via GitHub Issues
- **Documentation**: See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment guide
- **Health Check**: Monitor your deployment at `/health` endpoint

---

**Ready to command some incidents?** 🚨➡️🎯➡️✅

Start playing at **http://localhost:8080** (local) or deploy to your server with `make deploy-ubuntu`!