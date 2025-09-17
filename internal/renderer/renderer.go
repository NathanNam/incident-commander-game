package renderer

import (
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"github.com/NathanNam/incident-commander-game/internal/game"
)

// Renderer handles canvas drawing operations
type Renderer struct {
	canvas    js.Value
	ctx       js.Value
	cellSize  int
	mascotImg js.Value

	// Instrumentation fields
	renderCount       int64
	lastRenderTime    time.Time
	totalRenderTime   time.Duration
	lastMetricsReport time.Time
	imageLoadChecks   int64
	fallbackRenders   int64
}

// logRenderMetric logs a rendering metric to console for observability
func (r *Renderer) logRenderMetric(metric string, value interface{}, context string) {
	if js.Global().Get("console").Truthy() {
		js.Global().Get("console").Call("log",
			fmt.Sprintf("[RENDER_METRIC] %s: %v - Context: %s",
				metric, value, context))
	}
}

// New creates a new renderer
func New(canvas js.Value) *Renderer {
	ctx := canvas.Call("getContext", "2d")

	// Get the mascot image element
	document := js.Global().Get("document")
	mascotImg := document.Call("getElementById", "mascot-img")

	r := &Renderer{
		canvas:    canvas,
		ctx:       ctx,
		cellSize:  30, // Will be calculated dynamically
		mascotImg: mascotImg,

		// Initialize instrumentation
		renderCount:       0,
		lastRenderTime:    time.Now(),
		totalRenderTime:   0,
		lastMetricsReport: time.Now(),
		imageLoadChecks:   0,
		fallbackRenders:   0,
	}

	r.logRenderMetric("renderer_created", fmt.Sprintf("Canvas: %dx%d",
		canvas.Get("width").Int(), canvas.Get("height").Int()), "New renderer instance")

	return r
}

// updateCellSize calculates the cell size based on current canvas dimensions and grid size
func (r *Renderer) updateCellSize(g *game.Game) {
	canvasWidth := r.canvas.Get("width").Int()
	canvasHeight := r.canvas.Get("height").Int()

	// Use the smaller dimension to ensure the grid fits
	canvasSize := canvasWidth
	if canvasHeight < canvasWidth {
		canvasSize = canvasHeight
	}

	// Calculate cell size based on grid dimensions
	gridWidth := g.GetWidth()
	gridHeight := g.GetHeight()
	gridSize := gridWidth
	if gridHeight > gridWidth {
		gridSize = gridHeight
	}

	r.cellSize = canvasSize / gridSize
}

// Render renders the current game state
func (r *Renderer) Render(g *game.Game) {
	renderStart := time.Now()

	// Update cell size based on current canvas dimensions
	r.updateCellSize(g)

	r.clearCanvas()
	r.drawGrid(g)
	r.drawObstacles(g)
	r.drawTrail(g)
	r.drawAlerts(g)
	r.drawCommander(g)
	r.drawUI(g)

	// Track rendering metrics
	renderDuration := time.Since(renderStart)
	r.renderCount++
	r.totalRenderTime += renderDuration
	r.lastRenderTime = time.Now()

	// Report metrics every 5 seconds
	if time.Since(r.lastMetricsReport) >= 5*time.Second {
		avgRenderTime := r.totalRenderTime / time.Duration(r.renderCount)
		r.logRenderMetric("render_performance",
			fmt.Sprintf("Avg: %.2fms, Total renders: %d",
				float64(avgRenderTime.Nanoseconds())/1e6, r.renderCount),
			"Rendering performance stats")
		r.lastMetricsReport = time.Now()
	}
}

// clearCanvas clears the entire canvas
func (r *Renderer) clearCanvas() {
	// Set background color
	r.ctx.Set("fillStyle", "#1a1f36")
	r.ctx.Call("fillRect", 0, 0, r.canvas.Get("width"), r.canvas.Get("height"))
}

// drawGrid draws the game grid
func (r *Renderer) drawGrid(g *game.Game) {
	r.ctx.Set("strokeStyle", "#2a3f5f")
	r.ctx.Set("lineWidth", 1)

	width := g.GetWidth()
	height := g.GetHeight()

	// Draw vertical lines
	for x := 0; x <= width; x++ {
		r.ctx.Call("beginPath")
		r.ctx.Call("moveTo", x*r.cellSize, 0)
		r.ctx.Call("lineTo", x*r.cellSize, height*r.cellSize)
		r.ctx.Call("stroke")
	}

	// Draw horizontal lines
	for y := 0; y <= height; y++ {
		r.ctx.Call("beginPath")
		r.ctx.Call("moveTo", 0, y*r.cellSize)
		r.ctx.Call("lineTo", width*r.cellSize, y*r.cellSize)
		r.ctx.Call("stroke")
	}
}

// drawCommander draws the incident commander using the mascot image
func (r *Renderer) drawCommander(g *game.Game) {
	commander := g.GetCommander()
	x := commander.X * r.cellSize
	y := commander.Y * r.cellSize

	// Check if image is loaded and valid
	r.imageLoadChecks++
	imageLoaded := false
	if !r.mascotImg.IsNull() {
		// Check if image is actually loaded (complete property)
		complete := r.mascotImg.Get("complete")
		naturalWidth := r.mascotImg.Get("naturalWidth")

		// Image is considered loaded if complete=true and naturalWidth>0
		if !complete.IsUndefined() && complete.Bool() &&
			!naturalWidth.IsUndefined() && naturalWidth.Int() > 0 {
			imageLoaded = true
		}
	}

	if imageLoaded {
		// Draw the mascot image
		r.ctx.Call("drawImage", r.mascotImg, x, y, r.cellSize, r.cellSize)
	} else {
		// Fallback: draw a blue circle with eyes (o11y style)
		r.fallbackRenders++

		// Log fallback usage periodically
		if r.fallbackRenders == 1 {
			r.logRenderMetric("mascot_fallback", "first_use", "Using fallback graphics instead of mascot image")
		} else if r.fallbackRenders%100 == 0 {
			r.logRenderMetric("mascot_fallback", r.fallbackRenders, "Continued fallback rendering")
		}

		r.ctx.Set("fillStyle", "#9dd9f3")
		r.ctx.Call("beginPath")
		r.ctx.Call("arc", x+r.cellSize/2, y+r.cellSize/2, r.cellSize/2-2, 0, 2*3.14159)
		r.ctx.Call("fill")

		// Draw eyes
		r.ctx.Set("fillStyle", "#000000")
		eyeSize := r.cellSize / 8
		eyeOffsetX := r.cellSize / 4
		eyeOffsetY := r.cellSize / 3

		// Left eye
		r.ctx.Call("beginPath")
		r.ctx.Call("arc", x+r.cellSize/2-eyeOffsetX, y+eyeOffsetY, eyeSize, 0, 2*3.14159)
		r.ctx.Call("fill")

		// Right eye
		r.ctx.Call("beginPath")
		r.ctx.Call("arc", x+r.cellSize/2+eyeOffsetX, y+eyeOffsetY, eyeSize, 0, 2*3.14159)
		r.ctx.Call("fill")
	}

	// Log image load check metrics every 1000 checks
	if r.imageLoadChecks%1000 == 0 {
		r.logRenderMetric("image_load_checks", r.imageLoadChecks,
			fmt.Sprintf("Image loaded: %t, fallback renders: %d", imageLoaded, r.fallbackRenders))
	}
}

// drawTrail draws the commander's trail
func (r *Renderer) drawTrail(g *game.Game) {
	r.ctx.Set("fillStyle", "#6fcf3f")

	trail := g.GetTrail()
	for _, segment := range trail {
		x := segment.X * r.cellSize
		y := segment.Y * r.cellSize
		r.ctx.Call("fillRect", x+2, y+2, r.cellSize-4, r.cellSize-4)
	}
}

// drawAlerts draws the alert bubbles
func (r *Renderer) drawAlerts(g *game.Game) {
	alerts := g.GetAlerts()

	for _, alert := range alerts {
		x := alert.X * r.cellSize
		y := alert.Y * r.cellSize
		centerX := x + r.cellSize/2
		centerY := y + r.cellSize/2

		// Draw red circle
		r.ctx.Set("fillStyle", "#ff3838")
		r.ctx.Call("beginPath")
		r.ctx.Call("arc", centerX, centerY, r.cellSize/2-3, 0, 2*3.14159)
		r.ctx.Call("fill")

		// Draw exclamation mark
		r.ctx.Set("fillStyle", "#ffffff")
		r.ctx.Set("font", strconv.Itoa(r.cellSize/2)+"px Arial")
		r.ctx.Set("textAlign", "center")
		r.ctx.Set("textBaseline", "middle")
		r.ctx.Call("fillText", "!", centerX, centerY)
	}
}

// drawObstacles draws the level obstacles
func (r *Renderer) drawObstacles(g *game.Game) {
	r.ctx.Set("fillStyle", "#444444")

	obstacles := g.GetObstacles()
	for _, obstacle := range obstacles {
		x := obstacle.X * r.cellSize
		y := obstacle.Y * r.cellSize
		r.ctx.Call("fillRect", x, y, r.cellSize, r.cellSize)
	}
}

// drawUI draws the user interface elements
func (r *Renderer) drawUI(g *game.Game) {
	// Update DOM elements instead of drawing on canvas
	document := js.Global().Get("document")
	uiUpdates := 0

	// Update score
	scoreEl := document.Call("getElementById", "score")
	if !scoreEl.IsNull() {
		scoreEl.Set("textContent", "Score: "+strconv.Itoa(g.GetScore()))
		uiUpdates++
	}

	// Update level
	levelEl := document.Call("getElementById", "level")
	if !levelEl.IsNull() {
		levelEl.Set("textContent", "Level: "+strconv.Itoa(g.GetLevel()))
		uiUpdates++
	}

	// Update alerts progress
	alertsEl := document.Call("getElementById", "alerts")
	if !alertsEl.IsNull() {
		alertsText := "Alerts: " + strconv.Itoa(g.GetAlertsCollected()) + "/" + strconv.Itoa(g.GetAlertsNeeded())
		alertsEl.Set("textContent", alertsText)
		uiUpdates++
	}

	// Update game state
	stateEl := document.Call("getElementById", "game-state")
	if !stateEl.IsNull() {
		switch g.GetState() {
		case 0: // Playing
			stateEl.Set("textContent", "🎮 Playing")
			stateEl.Set("className", "playing")
		case 1: // Paused
			stateEl.Set("textContent", "⏸️ Paused")
			stateEl.Set("className", "paused")
		case 2: // GameOver
			stateEl.Set("textContent", "💀 Game Over")
			stateEl.Set("className", "game-over")
		case 3: // LevelComplete
			message := "🎉 Level " + strconv.Itoa(g.GetLevel()) + " Complete!"
			if g.GetLevel() < 10 {
				message += " → Level " + strconv.Itoa(g.GetLevel()+1)
			}
			stateEl.Set("textContent", message)
			stateEl.Set("className", "level-complete")
		}
		uiUpdates++
	}

	// Log UI update metrics every 500 renders
	if r.renderCount%500 == 0 {
		r.logRenderMetric("ui_updates", uiUpdates,
			fmt.Sprintf("DOM elements updated per render, total renders: %d", r.renderCount))
	}
}
