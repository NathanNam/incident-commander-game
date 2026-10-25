// Command game is the WebAssembly entry point. Go owns the rules and the canvas;
// the page (web/index.html) owns the controls, the overlays, the sound and the
// saved high score. They talk through window.ic:
//
//	page → Go:  ic.control("up"|"down"|"left"|"right"|"start"|"pause"|"restart"), ic.redraw()
//	Go → page:  ic.onFrame({state, score, level, alerts, needed}), ic.onEvent(name)
package main

import (
	"syscall/js"

	"github.com/nathannam/incident-commander-game/internal/game"
	"github.com/nathannam/incident-commander-game/internal/renderer"
)

var stateNames = map[game.State]string{
	game.Ready:         "ready",
	game.Playing:       "playing",
	game.Paused:        "paused",
	game.GameOver:      "gameover",
	game.LevelComplete: "levelclear",
	game.Won:           "won",
}

var eventNames = map[game.Event]string{
	game.EventCollect:  "collect",
	game.EventLevelUp:  "levelup",
	game.EventGameOver: "gameover",
	game.EventWon:      "won",
}

func main() {
	window := js.Global()
	canvas := window.Get("document").Call("getElementById", "game-canvas")
	ic := window.Get("ic")

	g := game.New(20, 20)
	r := renderer.New(canvas)

	draw := func() {
		r.Render(g)
		for _, e := range g.Events() {
			ic.Call("onEvent", eventNames[e])
		}
		ic.Call("onFrame", map[string]any{
			"state":  stateNames[g.State],
			"score":  g.Score,
			"level":  g.Level,
			"alerts": g.AlertsCollected,
			"needed": g.AlertsNeeded,
		})
	}

	ic.Set("control", js.FuncOf(func(this js.Value, args []js.Value) any {
		before := g.State
		switch args[0].String() {
		case "up":
			g.Steer(game.Up)
		case "down":
			g.Steer(game.Down)
		case "left":
			g.Steer(game.Left)
		case "right":
			g.Steer(game.Right)
		case "start":
			g.Start()
		case "pause":
			g.Pause()
		case "restart":
			g.Restart()
		}
		if g.State != before {
			draw()
		}
		return nil
	}))
	ic.Set("redraw", js.FuncOf(func(this js.Value, args []js.Value) any {
		draw()
		return nil
	}))

	// One move per tick; the tick gets shorter on later levels. The board is
	// redrawn every tick so the siren keeps blinking.
	var last float64
	var loop js.Func
	loop = js.FuncOf(func(this js.Value, args []js.Value) any {
		now := args[0].Float()
		if now-last >= float64(game.TickInterval(g.Level).Milliseconds()) {
			last = now
			g.Update()
			draw()
		}
		window.Call("requestAnimationFrame", loop)
		return nil
	})
	draw()
	ic.Call("onReady")
	window.Call("requestAnimationFrame", loop)

	select {}
}
