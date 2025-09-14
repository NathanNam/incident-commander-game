# Incident Commander Game - Product Requirements Document

## 1. Executive Summary

**Product Name:** Incident Commander  
**Type:** Browser-based Snake-like game  
**Platform:** Web (Desktop & Mobile Chrome)  
**Technology:** Go (with WebAssembly for browser compatibility)  
**Theme:** Incident management and observability (o11y)

## 2. Game Overview

### 2.1 Concept
Players control an "Incident Commander" (represented by the o11y mascot) that moves around the game field collecting alert notifications. As alerts are collected, the commander grows a trail of resolved incidents behind them, similar to the classic Snake game mechanic.

### 2.2 Core Gameplay Loop
1. Player controls the Incident Commander's movement direction
2. Commander automatically moves forward continuously
3. Collect red alert bubbles to score points and grow the trail
4. Avoid colliding with walls and your own trail
5. Complete level objectives to progress through 10 levels

## 3. Technical Requirements

### 3.1 Technology Stack
- **Backend:** Go 1.21+
- **Frontend Compilation:** WebAssembly (WASM)
- **Build System:** Go build with WASM target
- **Web Server:** Simple HTTP server to serve static files
- **Graphics:** HTML5 Canvas API via Go's syscall/js package
- **Browser Support:** Chrome (desktop and mobile), with fallback for other modern browsers

### 3.2 Architecture
```
incident-commander/
├── cmd/
│   ├── server/        # HTTP server (main.go)
│   └── game/          # WASM game entry (main.go)
├── internal/
│   ├── game/          # Core game logic
│   ├── renderer/      # Canvas rendering logic
│   ├── input/         # Input handling (keyboard/touch)
│   └── assets/        # Game assets (embedded)
├── web/               # Static web files
│   ├── index.html     # Game HTML wrapper
│   ├── style.css      # Game styling
│   ├── game.wasm      # Compiled WASM (generated)
│   └── wasm_exec.js   # Go WASM support (generated)
├── Makefile           # Build automation
├── go.mod             # Go module definition
└── README.md          # Project documentation
```

### 3.3 Session Management
- Each browser session creates a new game instance
- No persistent storage required (localStorage optional for high scores)
- Game state maintained in memory during session

## 4. Game Mechanics

### 4.1 Character Design
- **Incident Commander:** The o11y mascot (light blue circular character with eyes)
- **Trail:** Resolved incident tickets following the commander (green checkmarks or resolved ticket icons)
- **Collectibles:** Red alert bubbles with exclamation marks
- **Power-ups:** Special items (optional) like "Auto-resolve" or "Speed boost"

### 4.2 Movement System
- **Desktop:** Arrow keys or WASD for direction control
- **Mobile:** Swipe gestures or on-screen directional buttons
- **Speed:** Increases progressively with level
- **Grid-based:** Movement on invisible grid (20x20 default, scalable)

### 4.3 Collision System
- Wall collision: Game over
- Self-collision: Game over
- Alert collision: Score increase, trail growth
- Power-up collision: Temporary effect activation

### 4.4 Scoring System
- Base points per alert: 10 points
- Combo multiplier: Consecutive alerts without near-misses
- Level completion bonus: 100 * level number
- Time bonus: Extra points for quick level completion

## 5. Level Design

### 5.1 Level Progression (10 Levels)

**Level 1: "First Incident"**
- Grid: 20x20
- Speed: Slow (200ms/move)
- Alerts needed: 5
- Obstacles: None

**Level 2: "Peak Hours"**
- Grid: 20x20
- Speed: Slow-Medium (175ms/move)
- Alerts needed: 8
- Obstacles: None

**Level 3: "System Boundaries"**
- Grid: 22x22
- Speed: Medium (150ms/move)
- Alerts needed: 10
- Obstacles: 2 static barriers

**Level 4: "Service Mesh"**
- Grid: 24x24
- Speed: Medium (150ms/move)
- Alerts needed: 12
- Obstacles: 4 static barriers in cross pattern

**Level 5: "Cascade Failure"**
- Grid: 24x24
- Speed: Medium-Fast (125ms/move)
- Alerts needed: 15
- Obstacles: Moving obstacle (slow patrol)

**Level 6: "Load Balancer"**
- Grid: 26x26
- Speed: Fast (100ms/move)
- Alerts needed: 18
- Obstacles: 2 moving obstacles

**Level 7: "Circuit Breaker"**
- Grid: 26x26
- Speed: Fast (100ms/move)
- Alerts needed: 20
- Obstacles: Appearing/disappearing walls

**Level 8: "Distributed Chaos"**
- Grid: 28x28
- Speed: Very Fast (75ms/move)
- Alerts needed: 22
- Obstacles: Random barrier spawns

**Level 9: "Critical Path"**
- Grid: 30x30
- Speed: Very Fast (75ms/move)
- Alerts needed: 25
- Obstacles: Maze-like layout

**Level 10: "Incident Storm"**
- Grid: 30x30
- Speed: Extreme (50ms/move)
- Alerts needed: 30
- Obstacles: All previous mechanics combined

## 6. User Interface

### 6.1 Game Screen Layout
```
+---------------------------+
|  Score: 1250  Level: 3   |
|  Alerts: 5/10  Time: 45s  |
+---------------------------+
|                           |
|      GAME CANVAS          |
|      (Responsive)         |
|    [o11y mascot image]    |
|                           |
+---------------------------+
|  [↑]  [PAUSE]  [RESTART]  |
|  [←][→]     (Mobile)      |
|  [↓]                      |
+---------------------------+
```

### 6.1.1 HTML Structure
```html
<!DOCTYPE html>
<html>
<head>
    <title>Incident Commander</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <!-- Preload mascot image -->
    <img id="mascot-img" src="/images/o11y_alert.png" style="display:none;">
    
    <!-- Game UI -->
    <div id="game-container">
        <div id="score-panel">
            <span id="score">Score: 0</span>
            <span id="level">Level: 1</span>
            <span id="alerts">Alerts: 0/5</span>
        </div>
        <canvas id="game-canvas"></canvas>
        <div id="controls" class="mobile-only">
            <!-- Touch controls -->
        </div>
    </div>
    
    <script src="wasm_exec.js"></script>
    <script src="game.js"></script>
</body>
</html>
```

### 6.2 Visual Design
- **Color Palette:**
  - Background: Dark blue (#1a1f36)
  - Grid lines: Subtle gray (#2a3f5f)
  - Commander: Light blue (#9dd9f3)
  - Alerts: Red (#ff3838)
  - Trail: Green (#6fcf3f)
  - UI Text: White (#ffffff)

### 6.3 Responsive Design
- Canvas scales to fit viewport
- Minimum size: 320x320px
- Maximum size: 800x800px
- Touch controls appear only on mobile
- Font sizes scale with viewport

## 7. Audio (Optional Enhancement)

### 7.1 Sound Effects
- Alert collection: "ding" sound
- Level complete: Success fanfare
- Game over: Error sound
- Background: Subtle tech ambiance

## 8. Implementation Specifications

### 8.1 Main Game Loop
```go
type Game struct {
    Level         int
    Score         int
    Grid          [][]Cell
    Commander     Position
    Trail         []Position
    Alerts        []Position
    Direction     Direction
    GameState     State
    LastUpdate    time.Time
}

// Core game loop (60 FPS target)
func (g *Game) Update() {
    g.HandleInput()
    g.MoveCommander()
    g.CheckCollisions()
    g.SpawnAlerts()
    g.CheckLevelComplete()
    g.Render()
}
```

### 8.2 WASM Bridge
```go
// JavaScript interop for canvas rendering
js.Global().Get("canvas").Call("getContext", "2d")
// Event listeners for input
js.Global().Call("addEventListener", "keydown", callback)
```

### 8.3 Build Configuration
```makefile
# Build WASM
GOOS=js GOARCH=wasm go build -o web/wasm/game.wasm cmd/game/main.go

# Copy WASM support file
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/static/

# Run server
go run cmd/server/main.go
```

## 9. Performance Requirements

### 9.1 Target Metrics
- Load time: < 3 seconds on 3G connection
- Frame rate: 60 FPS on modern devices
- WASM size: < 5MB compressed
- Memory usage: < 50MB
- Battery efficient on mobile

### 9.2 Optimization Strategies
- Efficient collision detection (spatial hashing)
- Minimal DOM manipulation
- RequestAnimationFrame for rendering
- Object pooling for alerts
- Lazy loading of level assets

## 10. Testing Requirements

### 10.1 Browser Testing
- Chrome Desktop (latest 3 versions)
- Chrome Mobile (Android/iOS)
- Safari (fallback support)
- Firefox (fallback support)

### 10.2 Device Testing
- Desktop: 1920x1080, 1366x768
- Tablet: iPad, Android tablets
- Mobile: iPhone 12+, Pixel 5+

### 10.3 Gameplay Testing
- All 10 levels completable
- Controls responsive (< 50ms input lag)
- No memory leaks over extended play
- Consistent frame rate

## 11. Future Enhancements (Post-MVP)

- Multiplayer mode (WebSocket-based)
- Leaderboard system
- Custom level editor
- Achievement system
- More power-ups and obstacles
- Theme customization
- Sound effects and music
- Progressive Web App (PWA) support
- Save game state

## 12. Success Metrics

- Game loads and runs smoothly on target browsers
- All 10 levels are playable and winnable
- Intuitive controls on both desktop and mobile
- No critical bugs or crashes
- Engaging gameplay loop that encourages replay

## 13. Development Milestones

### Phase 1: Core Engine (Week 1)
- Basic Go/WASM setup
- Game loop implementation
- Canvas rendering system
- Input handling

### Phase 2: Game Mechanics (Week 2)
- Commander movement
- Collision detection
- Alert spawning
- Score system

### Phase 3: Level System (Week 3)
- Level progression
- Obstacle implementation
- Difficulty scaling
- Win/lose conditions

### Phase 4: Polish & Testing (Week 4)
- UI improvements
- Mobile optimization
- Bug fixes
- Performance optimization

## 14. Asset Requirements

### 14.1 Graphics (SVG or Canvas-drawn)
- O11y mascot sprite
- Alert bubble icon
- Trail segment graphics
- UI elements (buttons, score display)
- Background patterns

### 14.2 Fonts
- Primary: System font stack (sans-serif)
- Monospace for scores/timers

## 15. Deployment

### 15.1 Hosting Requirements
- Static file hosting (GitHub Pages, Netlify, or similar)
- CORS headers for WASM
- HTTPS required for mobile features
- CDN for global distribution

### 15.2 File Structure
```
/index.html         # Game entry point
/game.wasm         # Compiled game logic
/wasm_exec.js      # Go WASM support
/style.css         # Game styling
/manifest.json     # PWA manifest (future)