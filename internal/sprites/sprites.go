// Package sprites holds the game's pixel art. The art lives in sprites.json so the
// brand script (scripts/brand/make.mjs) can draw the favicon and share image from
// the same source.
package sprites

import (
	_ "embed"
	"syscall/js"
)

//go:embed sprites.json
var raw []byte

// Sprite is a grid of palette keys, one string per row; '.' is transparent.
type Sprite []string

// Palette maps a sprite letter to a CSS colour.
type Palette map[string]string

// Beacon is the colour of the light on the hero's head.
type Beacon string

const (
	BeaconAlert Beacon = "alert" // red: alerts are firing
	BeaconDim   Beacon = "dim"   // the red light's off frame, and knocked out
	BeaconCalm  Beacon = "calm"  // green: level cleared
	BeaconReady Beacon = "ready" // amber: waiting to start
)

var (
	hero, heroOuch, alert, check Sprite
	alertPal, checkPal           Palette
	heros                        = map[Beacon]Palette{}
)

// init parses sprites.json with the browser's JSON.parse; Go's encoding/json
// would more than double the size of game.wasm.
func init() {
	v := js.Global().Get("JSON").Call("parse", string(raw))
	hero, heroOuch = sprite(v.Get("hero")), sprite(v.Get("heroOuch"))
	alert, check = sprite(v.Get("alert")), sprite(v.Get("check"))
	pal := v.Get("palette")
	alertPal, checkPal = palette(pal.Get("alert")), palette(pal.Get("check"))
	base := palette(pal.Get("hero"))
	for _, b := range []Beacon{BeaconAlert, BeaconDim, BeaconCalm, BeaconReady} {
		p := Palette{}
		for k, c := range base {
			p[k] = c
		}
		for k, c := range palette(pal.Get("beacon").Get(string(b))) {
			p[k] = c
		}
		heros[b] = p
	}
}

func sprite(v js.Value) Sprite {
	s := make(Sprite, v.Length())
	for i := range s {
		s[i] = v.Index(i).String()
	}
	return s
}

func palette(v js.Value) Palette {
	p := Palette{}
	keys := js.Global().Get("Object").Call("keys", v)
	for i := 0; i < keys.Length(); i++ {
		k := keys.Index(i).String()
		p[k] = v.Get(k).String()
	}
	return p
}

// Hero returns the commander sprite with its beacon lit; ouch is the knocked-out face.
func Hero(b Beacon, ouch bool) (Sprite, Palette) {
	if ouch {
		return heroOuch, heros[b]
	}
	return hero, heros[b]
}

// Alert returns the warning-triangle sprite.
func Alert() (Sprite, Palette) { return alert, alertPal }

// Check returns the tick drawn on each resolved trail tile.
func Check() (Sprite, Palette) { return check, checkPal }
