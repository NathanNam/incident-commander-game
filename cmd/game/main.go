package main

import (
	"syscall/js"
	"time"

	"github.com/nathannam/incident-commander-game/internal/game"
	"github.com/nathannam/incident-commander-game/internal/renderer"
	"github.com/nathannam/incident-commander-game/internal/input"
)

func main() {
	println("🎮 Incident Commander WASM starting...")
	
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
		// Set error message in DOM
		document.Call("getElementById", "loading").Set("innerHTML", 
			"<div>❌ Canvas not found</div><div>Please refresh the page</div>")
		return
	}

	println("✅ Canvas found, initializing game...")

	// Initialize game components
	g := game.New(20, 20)
	r := renderer.New(canvas)
	inputHandler := input.New()

	println("✅ Game components initialized")

	// Set up event listeners
	inputHandler.SetupEventListeners(g)

	println("✅ Event listeners set up")

	// Initial render
	r.Render(g)

	println("✅ Initial render complete")

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
			// Always update to handle level transitions, but render depends on game state
			g.Update()
			r.Render(g)
			lastUpdate = now
		}
		
		// Continue the animation loop
		js.Global().Call("requestAnimationFrame", gameLoop)
		return nil
	})

	// Start the game loop
	js.Global().Call("requestAnimationFrame", gameLoop)
	
	println("✅ Game loop started!")
	println("🎮 Incident Commander is ready to play!")

	// Keep the program running - use a channel instead of select {}
	done := make(chan bool)
	<-done
}