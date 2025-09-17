package input

import (
	"fmt"
	"syscall/js"
	"time"

	"github.com/NathanNam/incident-commander-game/internal/game"
)

// InputHandler manages input events
type InputHandler struct {
	keyCallback              js.Func
	touchStartCallback       js.Func
	touchEndCallback         js.Func
	touchStartX, touchStartY float64

	// Instrumentation fields
	keyPressCount     int64
	touchEventCount   int64
	swipeCount        int64
	tapCount          int64
	buttonPressCount  int64
	lastMetricsReport time.Time
}

// logInputMetric logs an input metric to console for observability
func (h *InputHandler) logInputMetric(metric string, value interface{}, context string) {
	if js.Global().Get("console").Truthy() {
		js.Global().Get("console").Call("log",
			fmt.Sprintf("[INPUT_METRIC] %s: %v - Context: %s",
				metric, value, context))
	}
}

// New creates a new input handler
func New() *InputHandler {
	h := &InputHandler{
		keyPressCount:     0,
		touchEventCount:   0,
		swipeCount:        0,
		tapCount:          0,
		buttonPressCount:  0,
		lastMetricsReport: time.Now(),
	}

	h.logInputMetric("input_handler_created", "initialized", "New input handler instance")
	return h
}

// SetupEventListeners sets up keyboard and touch event listeners
func (h *InputHandler) SetupEventListeners(g *game.Game) {
	document := js.Global().Get("document")

	// Keyboard events
	h.keyCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		key := event.Get("key").String()
		h.keyPressCount++

		// Log key press
		h.logInputMetric("key_press", key, fmt.Sprintf("Total key presses: %d", h.keyPressCount))

		switch key {
		case "ArrowUp", "w", "W":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(0)) // Up
			h.logInputMetric("direction_input", "up", "Keyboard direction change")
		case "ArrowDown", "s", "S":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(1)) // Down
			h.logInputMetric("direction_input", "down", "Keyboard direction change")
		case "ArrowLeft", "a", "A":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(2)) // Left
			h.logInputMetric("direction_input", "left", "Keyboard direction change")
		case "ArrowRight", "d", "D":
			event.Call("preventDefault")
			g.SetDirection(game.Direction(3)) // Right
			h.logInputMetric("direction_input", "right", "Keyboard direction change")
		case " ", "p", "P":
			event.Call("preventDefault")
			g.Pause()
			h.logInputMetric("game_control", "pause", "Keyboard pause/resume")
		case "r", "R":
			event.Call("preventDefault")
			g.Restart()
			h.logInputMetric("game_control", "restart", "Keyboard restart")
		}

		// Report metrics every 50 key presses
		if h.keyPressCount%50 == 0 {
			h.reportInputMetrics()
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
		h.touchEventCount++

		touches := event.Get("touches")
		if touches.Get("length").Int() > 0 {
			touch := touches.Index(0)
			h.touchStartX = touch.Get("clientX").Float()
			h.touchStartY = touch.Get("clientY").Float()

			h.logInputMetric("touch_start", fmt.Sprintf("(%.1f,%.1f)", h.touchStartX, h.touchStartY),
				fmt.Sprintf("Touch events: %d", h.touchEventCount))
		}

		return nil
	})

	// Touch end - determine swipe direction
	h.touchEndCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		h.touchEventCount++

		changedTouches := event.Get("changedTouches")
		if changedTouches.Get("length").Int() > 0 {
			touch := changedTouches.Index(0)
			endX := touch.Get("clientX").Float()
			endY := touch.Get("clientY").Float()

			deltaX := endX - h.touchStartX
			deltaY := endY - h.touchStartY
			minDistance := 30.0

			h.logInputMetric("touch_end", fmt.Sprintf("Delta: (%.1f,%.1f)", deltaX, deltaY),
				"Touch gesture analysis")

			// Determine swipe direction
			if abs(deltaX) > abs(deltaY) {
				// Horizontal swipe
				if abs(deltaX) > minDistance {
					h.swipeCount++
					if deltaX > 0 {
						g.SetDirection(game.Direction(3)) // Right
						h.logInputMetric("swipe_direction", "right", fmt.Sprintf("Swipes: %d", h.swipeCount))
					} else {
						g.SetDirection(game.Direction(2)) // Left
						h.logInputMetric("swipe_direction", "left", fmt.Sprintf("Swipes: %d", h.swipeCount))
					}
				}
			} else {
				// Vertical swipe
				if abs(deltaY) > minDistance {
					h.swipeCount++
					if deltaY > 0 {
						g.SetDirection(game.Direction(1)) // Down
						h.logInputMetric("swipe_direction", "down", fmt.Sprintf("Swipes: %d", h.swipeCount))
					} else {
						g.SetDirection(game.Direction(0)) // Up
						h.logInputMetric("swipe_direction", "up", fmt.Sprintf("Swipes: %d", h.swipeCount))
					}
				} else if abs(deltaX) < 10 && abs(deltaY) < 10 {
					// This was a tap, pause the game
					h.tapCount++
					g.Pause()
					h.logInputMetric("tap_gesture", "pause", fmt.Sprintf("Taps: %d", h.tapCount))
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
			dirName := []string{"up", "down", "left", "right"}[direction]
			callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				args[0].Call("preventDefault")
				h.buttonPressCount++
				g.SetDirection(game.Direction(direction))
				h.logInputMetric("button_press", dirName,
					fmt.Sprintf("Button presses: %d", h.buttonPressCount))
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
			h.buttonPressCount++
			g.Pause()
			h.logInputMetric("button_press", "pause",
				fmt.Sprintf("Button presses: %d", h.buttonPressCount))
			return nil
		})
		pauseBtn.Call("addEventListener", "touchstart", pauseCallback)
	}

	// Restart button
	restartBtn := document.Call("getElementById", "btn-restart")
	if !restartBtn.IsNull() {
		restartCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			args[0].Call("preventDefault")
			h.buttonPressCount++
			g.Restart()
			h.logInputMetric("button_press", "restart",
				fmt.Sprintf("Button presses: %d", h.buttonPressCount))
			return nil
		})
		restartBtn.Call("addEventListener", "touchstart", restartCallback)
	}
}

// reportInputMetrics reports comprehensive input metrics
func (h *InputHandler) reportInputMetrics() {
	now := time.Now()
	if now.Sub(h.lastMetricsReport) >= 5*time.Second {
		h.logInputMetric("input_summary",
			fmt.Sprintf("Keys: %d, Touch: %d, Swipes: %d, Taps: %d, Buttons: %d",
				h.keyPressCount, h.touchEventCount, h.swipeCount, h.tapCount, h.buttonPressCount),
			"Comprehensive input statistics")
		h.lastMetricsReport = now
	}
}

// Cleanup releases event listeners
func (h *InputHandler) Cleanup() {
	h.logInputMetric("input_cleanup", "releasing_callbacks", "Input handler cleanup")

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
