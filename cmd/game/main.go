package main

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/NathanNam/incident-commander-game/internal/game"
	"github.com/NathanNam/incident-commander-game/internal/input"
	"github.com/NathanNam/incident-commander-game/internal/renderer"
	"github.com/NathanNam/incident-commander-game/internal/telemetry"
)

// Game metrics and telemetry
var (
	gameStartTime   time.Time
	frameCount      int64
	lastFPSReport   time.Time
	gameEvents      []GameEvent
	clientTelemetry *telemetry.ClientTelemetry
)

// GameEvent represents a game event for telemetry
type GameEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Level     int       `json:"level,omitempty"`
	Score     int       `json:"score,omitempty"`
	Data      string    `json:"data,omitempty"`
}

// logGameEvent logs a game event to console and stores it
func logGameEvent(eventType string, level, score int, data string) {
	event := GameEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Level:     level,
		Score:     score,
		Data:      data,
	}
	gameEvents = append(gameEvents, event)

	// Log to browser console for observability
	js.Global().Get("console").Call("log",
		fmt.Sprintf("[GAME_EVENT] %s - Level: %d, Score: %d, Data: %s",
			eventType, level, score, data))

	// Send to telemetry system with enhanced attributes
	if clientTelemetry != nil {
		attributes := map[string]interface{}{
			"timestamp": event.Timestamp.Format(time.RFC3339),
			"source":    "client_game",
		}

		// Add performance context
		if eventType == "level_change" || eventType == "score_change" {
			attributes["game_duration_seconds"] = time.Since(gameStartTime).Seconds()
			attributes["frames_rendered"] = frameCount
		}

		clientTelemetry.LogEvent(eventType, level, score, data, attributes)
	}
}

// reportPerformanceMetrics reports FPS and other performance metrics
func reportPerformanceMetrics(currentLevel int) {
	now := time.Now()
	if now.Sub(lastFPSReport) >= 5*time.Second {
		elapsed := now.Sub(lastFPSReport)
		fps := float64(frameCount) / elapsed.Seconds()

		js.Global().Get("console").Call("log",
			fmt.Sprintf("[PERFORMANCE] FPS: %.2f, Level: %d, Frames: %d",
				fps, currentLevel, frameCount))

		// Send performance metrics to telemetry
		if clientTelemetry != nil {
			clientTelemetry.RecordMetric("fps", fps, "gauge", map[string]interface{}{
				"level": currentLevel,
				"game_duration_seconds": time.Since(gameStartTime).Seconds(),
			})

			clientTelemetry.RecordMetric("frames_rendered_total", float64(frameCount), "counter", map[string]interface{}{
				"level": currentLevel,
			})
		}

		frameCount = 0
		lastFPSReport = now
	}
}

// getCurrentServerURL gets the current server URL from the browser
func getCurrentServerURL() string {
	location := js.Global().Get("location")
	protocol := location.Get("protocol").String()
	hostname := location.Get("hostname").String()
	port := location.Get("port").String()

	if port != "" && port != "80" && port != "443" {
		return protocol + "//" + hostname + ":" + port
	}
	return protocol + "//" + hostname
}

func main() {
	gameStartTime = time.Now()
	lastFPSReport = time.Now()

	// Initialize client telemetry system
	serverURL := getCurrentServerURL()
	clientTelemetry = telemetry.NewClientTelemetry(serverURL)

	println("🎮 Incident Commander WASM starting...")
	logGameEvent("game_start", 0, 0, "WebAssembly initialization")

	// Wait for the DOM to be ready
	document := js.Global().Get("document")

	// Wait for canvas to be available
	var canvas js.Value
	for i := 0; i < 50; i++ { // Try for 5 seconds
		canvas = document.Call("getElementById", "game-canvas")
		if !canvas.IsNull() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if canvas.IsNull() {
		println("❌ Canvas element not found after waiting")
		logGameEvent("error", 0, 0, "Canvas element not found")
		// Set error message in DOM
		document.Call("getElementById", "loading").Set("innerHTML",
			"<div>❌ Canvas not found</div><div>Please refresh the page</div>")
		return
	}

	println("✅ Canvas found, initializing game...")
	logGameEvent("canvas_found", 0, 0, "Canvas element located successfully")

	// Initialize game components with tracing
	initSpan := clientTelemetry.StartSpan("game_initialization")
	initSpan.SetAttribute("canvas_width", canvas.Get("width").Int())
	initSpan.SetAttribute("canvas_height", canvas.Get("height").Int())

	g := game.New(20, 20)
	r := renderer.New(canvas)
	inputHandler := input.New()

	initSpan.End()

	println("✅ Game components initialized")
	logGameEvent("components_initialized", 1, 0, "Game, renderer, and input handler created")

	// Set up event listeners
	inputHandler.SetupEventListeners(g)

	println("✅ Event listeners set up")
	logGameEvent("event_listeners_setup", 1, 0, "Input event listeners configured")

	// Initial render
	r.Render(g)

	println("✅ Initial render complete")
	logGameEvent("initial_render", 1, 0, "First game frame rendered")

	// Game loop using requestAnimationFrame for better performance
	var gameLoop js.Func
	var lastUpdate float64

	// Better speed progression - faster but still playable
	getTargetFPS := func(level int) float64 {
		// Level 1: 2 FPS (500ms), Level 10: 8 FPS (125ms)
		fps := 1.5 + float64(level)*0.65 // 2.15 to 8 FPS range
		if fps > 8 {
			fps = 8 // Maximum 8 FPS
		}
		return fps
	}

	gameLoop = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		now := args[0].Float()
		targetFPS := getTargetFPS(g.GetLevel())

		if now-lastUpdate >= 1000.0/targetFPS {
			// Start game loop span for performance tracking
			loopSpan := clientTelemetry.StartSpan("game_loop_iteration")
			loopSpan.SetAttribute("target_fps", targetFPS)
			loopSpan.SetAttribute("frame_number", frameCount)

			// Track previous state for event detection
			prevLevel := g.GetLevel()
			prevScore := g.GetScore()
			prevState := g.GetState()

			// Always update to handle level transitions, but render depends on game state
			g.Update()
			r.Render(g)
			lastUpdate = now
			frameCount++

			// End the game loop span
			loopSpan.SetAttribute("level", g.GetLevel())
			loopSpan.SetAttribute("score", g.GetScore())
			loopSpan.End()

			// Check for state changes and log events
			currentLevel := g.GetLevel()
			currentScore := g.GetScore()
			currentState := g.GetState()

			// Log level changes
			if currentLevel != prevLevel {
				logGameEvent("level_change", currentLevel, currentScore,
					fmt.Sprintf("Advanced from level %d to %d", prevLevel, currentLevel))
			}

			// Log score changes (alert collection)
			if currentScore != prevScore {
				logGameEvent("score_change", currentLevel, currentScore,
					fmt.Sprintf("Score increased by %d", currentScore-prevScore))
			}

			// Log game state changes
			if currentState != prevState {
				stateStr := map[int]string{0: "Playing", 1: "Paused", 2: "GameOver", 3: "LevelComplete"}
				logGameEvent("state_change", currentLevel, currentScore,
					fmt.Sprintf("State changed to %s", stateStr[int(currentState)]))
			}

			// Report performance metrics periodically
			reportPerformanceMetrics(currentLevel)
		}

		// Continue the animation loop
		js.Global().Call("requestAnimationFrame", gameLoop)
		return nil
	})

	// Start the game loop
	js.Global().Call("requestAnimationFrame", gameLoop)

	println("✅ Game loop started!")
	println("🎮 Incident Commander is ready to play!")
	logGameEvent("game_ready", 1, 0, "Game loop started and ready for player input")

	// Log total initialization time
	initTime := time.Since(gameStartTime)
	logGameEvent("initialization_complete", 1, 0,
		fmt.Sprintf("Total initialization time: %v", initTime))

	// Keep the program running - use a channel instead of select {}
	done := make(chan bool)
	<-done
}
