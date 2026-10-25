// Package renderer draws the board on an HTML canvas through syscall/js.
package renderer

import (
	"syscall/js"

	"github.com/nathannam/incident-commander-game/internal/game"
	"github.com/nathannam/incident-commander-game/internal/sprites"
)

// Board colours (the page's CSS uses the same palette).
const (
	colBG     = "#0b0714"
	colGrid   = "#171026"
	colWall   = "#3a2a63"
	colWallHi = "#6f56c9"
	colTrail  = "#1aa77a"
	colTrailD = "#0f6b4e"
)

// Renderer handles canvas drawing operations
type Renderer struct {
	canvas js.Value
	ctx    js.Value
	cell   float64
	frame  int
}

// New creates a new renderer
func New(canvas js.Value) *Renderer {
	return &Renderer{canvas: canvas, ctx: canvas.Call("getContext", "2d")}
}

// Render draws the whole board. The page sizes the canvas (in device pixels) to a
// multiple of the grid, so every cell lands on whole pixels.
func (r *Renderer) Render(g *game.Game) {
	r.frame++
	size := float64(r.canvas.Get("width").Int())
	r.cell = size / float64(g.Width)
	r.ctx.Set("imageSmoothingEnabled", false)

	r.fill(colBG, 0, 0, size, size)
	r.drawGrid(g, size)
	for _, o := range g.Obstacles {
		r.drawWall(o)
	}
	check, checkPal := sprites.Check()
	for _, t := range g.Trail {
		r.drawTrail(t, check, checkPal)
	}
	alert, alertPal := sprites.Alert()
	for _, a := range g.Alerts {
		r.sprite(alert, alertPal, float64(a.X)*r.cell+r.cell*0.03, float64(a.Y)*r.cell, r.cell*0.94)
	}
	r.drawCommander(g)
}

func (r *Renderer) fill(color string, x, y, w, h float64) {
	r.ctx.Set("fillStyle", color)
	r.ctx.Call("fillRect", x, y, w, h)
}

func (r *Renderer) drawGrid(g *game.Game, size float64) {
	r.ctx.Set("fillStyle", colGrid)
	for i := 1; i < g.Width; i++ {
		p := float64(i) * r.cell
		r.ctx.Call("fillRect", p, 0, 1, size)
		r.ctx.Call("fillRect", 0, p, size, 1)
	}
}

func (r *Renderer) drawWall(p game.Position) {
	x, y, c := float64(p.X)*r.cell, float64(p.Y)*r.cell, r.cell
	r.fill(colWall, x+1, y+1, c-2, c-2)
	r.fill(colWallHi, x+1, y+1, c-2, c/6)
}

// drawTrail draws a resolved incident: a green tile with a tick.
func (r *Renderer) drawTrail(p game.Position, check sprites.Sprite, pal sprites.Palette) {
	x, y, c := float64(p.X)*r.cell, float64(p.Y)*r.cell, r.cell
	pad := c * 0.1
	r.fill(colTrailD, x+pad, y+pad, c-2*pad, c-2*pad)
	r.fill(colTrail, x+pad, y+pad, c-2*pad, c-2*pad-c*0.08)
	r.sprite(check, pal, x+c*0.18, y+c*0.16, c*0.64)
}

// drawCommander draws the hero a little larger than its cell. Its beacon flashes
// red while alerts are firing, turns green on a cleared level and is amber on the
// start screen.
func (r *Renderer) drawCommander(g *game.Game) {
	beacon, ouch := sprites.BeaconAlert, false
	switch g.State {
	case game.Ready, game.Paused:
		beacon = sprites.BeaconReady
	case game.Playing:
		if r.frame%2 == 0 {
			beacon = sprites.BeaconDim
		}
	case game.LevelComplete, game.Won:
		beacon = sprites.BeaconCalm
	case game.GameOver:
		beacon, ouch = sprites.BeaconDim, true
	}
	p := g.Commander
	hero, pal := sprites.Hero(beacon, ouch)
	size := r.cell * 1.4
	r.sprite(hero, pal, float64(p.X)*r.cell-(size-r.cell)/2, float64(p.Y)*r.cell-(size-r.cell)*0.8, size)
}

// sprite draws pixel art scaled to size×size at (x, y).
func (r *Renderer) sprite(s sprites.Sprite, pal sprites.Palette, x, y, size float64) {
	px := size / float64(len(s))
	for j, row := range s {
		for i, ch := range row {
			color, ok := pal[string(ch)]
			if !ok {
				continue
			}
			r.ctx.Set("fillStyle", color)
			// Round outward so neighbouring pixels never leave hairline gaps.
			x0, y0 := float64(int(x+float64(i)*px)), float64(int(y+float64(j)*px))
			x1, y1 := float64(int(x+float64(i+1)*px+0.999)), float64(int(y+float64(j+1)*px+0.999))
			r.ctx.Call("fillRect", x0, y0, x1-x0, y1-y0)
		}
	}
}
