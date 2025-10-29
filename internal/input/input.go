package input

import (
	"syscall/js"

	"github.com/nathannam/incident-commander-game/internal/game"
)

// InputHandler manages input events
type InputHandler struct {
	keyCallback js.Func
	touchStartCallback js.Func
	touchEndCallback js.Func
	touchStartX, touchStartY float64
}

// New creates a new input handler
func New() *InputHandler {
	return &InputHandler{}
}

// SetupEventListeners sets up keyboard and touch event listeners
func (h *InputHandler) SetupEventListeners(g *game.Game) {
	document := js.Global().Get("document")

	// Keyboard events
	h.keyCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("key").String()
		
		switch key {
		case "ArrowUp", "w", "W":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(0)) // Up
		case "ArrowDown", "s", "S":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(1)) // Down
		case "ArrowLeft", "a", "A":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(2)) // Left
		case "ArrowRight", "d", "D":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(3)) // Right
		case " ", "p", "P":
			event.Call("preventDefault")
			g.Pause()
		case "r", "R":
			event.Call("preventDefault")
			g.Restart()
		}
		
		return nil
	})
	
	document.Call("addEventListener", "keydown", h.keyCallback)

	// Touch events for mobile
	h.setupTouchEvents(g)
}

// setupTouchEvents sets up touch events for mobile controls
func (h *InputHandler) setupTouchEvents(g *game.Game) {
	canvas := js.Global().Get("document").Call("getElementById", "game-canvas")
	
	// Touch start
	h.touchStartCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		
		touches := event.Get("touches")
		if touches.Get("length").Int() > 0 {
			touch := touches.Index(0)
			h.touchStartX = touch.Get("clientX").Float()
			h.touchStartY = touch.Get("clientY").Float()
		}
		
		return nil
	})
	
	// Touch end - determine swipe direction
	h.touchEndCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		
		changedTouches := event.Get("changedTouches")
		if changedTouches.Get("length").Int() > 0 {
			touch := changedTouches.Index(0)
			endX := touch.Get("clientX").Float()
			endY := touch.Get("clientY").Float()
			
			deltaX := endX - h.touchStartX
			deltaY := endY - h.touchStartY
			minDistance := 30.0
			
			// Determine swipe direction
			if abs(deltaX) > abs(deltaY) {
				// Horizontal swipe
				if abs(deltaX) > minDistance {
					if deltaX > 0 {
						g.SetDirection(game.Direction(3)) // Right
					} else {
						g.SetDirection(game.Direction(2)) // Left
					}
				}
			} else {
				// Vertical swipe
				if abs(deltaY) > minDistance {
					if deltaY > 0 {
						g.SetDirection(game.Direction(1)) // Down
					} else {
						g.SetDirection(game.Direction(0)) // Up
					}
				} else if abs(deltaX) < 10 && abs(deltaY) < 10 {
					// This was a tap, pause the game
					g.Pause()
				}
			}
		}
		
		return nil
	})
	
	canvas.Call("addEventListener", "touchstart", h.touchStartCallback)
	canvas.Call("addEventListener", "touchend", h.touchEndCallback)
	
	// Set up on-screen buttons if they exist
	h.setupOnScreenButtons(g)
}

// setupOnScreenButtons sets up on-screen button controls
func (h *InputHandler) setupOnScreenButtons(g *game.Game) {
	document := js.Global().Get("document")
	
	// Direction buttons
	buttons := []struct {
		id        string
		direction int
	}{
		{"btn-up", 0},
		{"btn-down", 1},
		{"btn-left", 2},
		{"btn-right", 3},
	}
	
	for _, btn := range buttons {
		element := document.Call("getElementById", btn.id)
		if !element.IsNull() {
			direction := btn.direction
			callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				args[0].Call("preventDefault")
				g.SetDirection(game.Direction(direction))
				return nil
			})
			element.Call("addEventListener", "touchstart", callback)
		}
	}
	
	// Pause button
	pauseBtn := document.Call("getElementById", "btn-pause")
	if !pauseBtn.IsNull() {
		pauseCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			args[0].Call("preventDefault")
			g.Pause()
			return nil
		})
		pauseBtn.Call("addEventListener", "touchstart", pauseCallback)
	}
	
	// Restart button
	restartBtn := document.Call("getElementById", "btn-restart")
	if !restartBtn.IsNull() {
		restartCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			args[0].Call("preventDefault")
			g.Restart()
			return nil
		})
		restartBtn.Call("addEventListener", "touchstart", restartCallback)
	}
}

// Cleanup releases event listeners
func (h *InputHandler) Cleanup() {
	if !h.keyCallback.IsUndefined() {
		h.keyCallback.Release()
	}
	if !h.touchStartCallback.IsUndefined() {
		h.touchStartCallback.Release()
	}
	if !h.touchEndCallback.IsUndefined() {
		h.touchEndCallback.Release()
	}
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}