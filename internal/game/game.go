// Package game holds the rules: movement, collisions, scoring and the ten levels.
// It has no browser dependencies, so it runs (and is tested) natively.
package game

import (
	"math/rand"
	"time"
)

// Direction represents movement direction
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

func (d Direction) opposite() Direction {
	switch d {
	case Up:
		return Down
	case Down:
		return Up
	case Left:
		return Right
	}
	return Left
}

// Position represents a coordinate on the game grid
type Position struct {
	X, Y int
}

// State represents the current state of the game
type State int

const (
	Ready         State = iota // start screen, waiting for the first move
	Playing                    // commander is moving
	Paused                     // paused by the player
	GameOver                   // hit a wall, an obstacle or the trail
	LevelComplete              // brief pause before the next level
	Won                        // cleared level 10
)

// MaxLevel is the last level.
const MaxLevel = 10

// levelPause is how long the "level clear" screen shows before the next level.
const levelPause = 1200 * time.Millisecond

// Event is something the page may want to react to (a sound, a high score).
type Event int

const (
	EventCollect Event = iota + 1
	EventLevelUp
	EventGameOver
	EventWon
)

// Game represents the main game structure
type Game struct {
	Width, Height     int
	Commander         Position
	Trail             []Position
	Alerts            []Position
	Obstacles         []Position
	Direction         Direction
	moved             Direction // direction of the last step, to block reversing into the trail
	State             State
	Score             int
	Level             int
	AlertsCollected   int
	AlertsNeeded      int
	StartTime         time.Time
	LevelCompleteTime time.Time
	events            []Event
	rng               *rand.Rand
	now               func() time.Time
}

// New creates a game on the start screen.
func New(width, height int) *Game {
	return newGame(width, height, rand.New(rand.NewSource(time.Now().UnixNano())), time.Now)
}

func newGame(width, height int, rng *rand.Rand, now func() time.Time) *Game {
	g := &Game{Width: width, Height: height, rng: rng, now: now}
	g.reset()
	return g
}

func (g *Game) reset() {
	g.State = Ready
	g.Score = 0
	g.Level = 1
	g.events = nil
	g.startLevel()
}

func (g *Game) startLevel() {
	g.Commander = Position{X: g.Width / 2, Y: g.Height / 2}
	g.Trail = nil
	g.Alerts = nil
	g.Obstacles = nil
	g.Direction, g.moved = Right, Right
	g.AlertsCollected = 0
	g.AlertsNeeded = alertsNeeded(g.Level)
	g.StartTime = g.now()
	g.setupLevel()
	g.spawnAlerts()
}

// alertsNeeded is the number of alerts to clear a level: 5 on level 1, 14 on level 10.
func alertsNeeded(level int) int { return 4 + level }

// TickInterval is the time between moves on a level: about 465 ms on level 1,
// down to 125 ms from level 10.
func TickInterval(level int) time.Duration {
	fps := 1.5 + float64(level)*0.65
	if fps > 8 {
		fps = 8
	}
	return time.Duration(float64(time.Second) / fps)
}

// Update advances the game by one tick.
func (g *Game) Update() {
	switch g.State {
	case LevelComplete:
		if g.now().Sub(g.LevelCompleteTime) >= levelPause {
			g.Level++
			g.startLevel()
			g.State = Playing
		}
	case Playing:
		g.step()
	}
}

func (g *Game) step() {
	next := g.Commander
	switch g.Direction {
	case Up:
		next.Y--
	case Down:
		next.Y++
	case Left:
		next.X--
	case Right:
		next.X++
	}
	// On a crash the commander stays where it was, so the board shows what it hit.
	if g.collides(next) {
		g.State = GameOver
		g.emit(EventGameOver)
		return
	}
	g.Trail = append(g.Trail, g.Commander)
	g.Commander = next
	g.moved = g.Direction
	for i, a := range g.Alerts {
		if a == g.Commander {
			g.collectAlert(i)
			break
		}
	}
}

func (g *Game) collides(p Position) bool {
	if p.X < 0 || p.X >= g.Width || p.Y < 0 || p.Y >= g.Height {
		return true
	}
	for _, s := range g.Trail {
		if s == p {
			return true
		}
	}
	for _, o := range g.Obstacles {
		if o == p {
			return true
		}
	}
	return false
}

// collectAlert scores an alert. Each one on a level is worth more than the last
// (10, 20, 30, ...), so a long run pays off.
func (g *Game) collectAlert(index int) {
	g.Alerts = append(g.Alerts[:index], g.Alerts[index+1:]...)
	g.AlertsCollected++
	g.Score += 10 * g.AlertsCollected

	if g.AlertsCollected < g.AlertsNeeded {
		g.spawnAlerts()
		g.emit(EventCollect)
		return
	}

	// Level clear: bonus for the level, plus up to 60 for finishing fast.
	timeBonus := max(0, 60-int(g.now().Sub(g.StartTime).Seconds()))
	g.Score += 100*g.Level + timeBonus
	if g.Level >= MaxLevel {
		g.State = Won
		g.emit(EventWon)
		return
	}
	g.State = LevelComplete
	g.LevelCompleteTime = g.now()
	g.emit(EventLevelUp)
}

func (g *Game) emit(e Event) { g.events = append(g.events, e) }

// Events returns and clears what happened since the last call.
func (g *Game) Events() []Event {
	ev := g.events
	g.events = nil
	return ev
}

// setupLevel configures obstacles and layout for the current level
func (g *Game) setupLevel() {
	switch g.Level {
	case 3, 4:
		g.addStaticBarriers()
	case 5, 6:
		g.addStaticBarriers()
		g.addRandomObstacles(2)
	case 7, 8:
		g.addRandomObstacles(4)
	case 9, 10:
		g.addMazeLayout()
	}
}

// spawnAlerts keeps three alerts on the board.
func (g *Game) spawnAlerts() {
	for len(g.Alerts) < 3 {
		pos := Position{X: g.rng.Intn(g.Width), Y: g.rng.Intn(g.Height)}
		if !g.isPositionOccupied(pos) {
			g.Alerts = append(g.Alerts, pos)
		}
	}
}

func (g *Game) isPositionOccupied(pos Position) bool {
	if g.Commander == pos {
		return true
	}
	for _, s := range g.Trail {
		if s == pos {
			return true
		}
	}
	for _, o := range g.Obstacles {
		if o == pos {
			return true
		}
	}
	for _, a := range g.Alerts {
		if a == pos {
			return true
		}
	}
	return false
}

// nearSpawn reports whether p is inside the 5×5 safe zone around the start.
func (g *Game) nearSpawn(p Position) bool {
	return abs(p.X-g.Width/2) <= 2 && abs(p.Y-g.Height/2) <= 2
}

// addStaticBarriers draws four crosses, leaving the middle clear.
func (g *Game) addStaticBarriers() {
	cx, cy := g.Width/2, g.Height/2
	for x := 2; x < g.Width-2; x++ {
		if abs(x-cx) > 2 {
			g.Obstacles = append(g.Obstacles, Position{X: x, Y: cy - 4}, Position{X: x, Y: cy + 4})
		}
	}
	for y := 2; y < g.Height-2; y++ {
		if abs(y-cy) > 2 {
			g.Obstacles = append(g.Obstacles, Position{X: cx - 4, Y: y}, Position{X: cx + 4, Y: y})
		}
	}
}

func (g *Game) addRandomObstacles(count int) {
	for i := 0; i < count; i++ {
		for attempts := 0; attempts < 50; attempts++ {
			pos := Position{X: g.rng.Intn(g.Width), Y: g.rng.Intn(g.Height)}
			if !g.nearSpawn(pos) && !g.isPositionOccupied(pos) {
				g.Obstacles = append(g.Obstacles, pos)
				break
			}
		}
	}
}

// addMazeLayout places short two-block walls on a grid.
func (g *Game) addMazeLayout() {
	for x := 2; x < g.Width-2; x += 4 {
		for y := 2; y < g.Height-2; y += 4 {
			pos := Position{X: x, Y: y}
			if g.nearSpawn(pos) || g.isPositionOccupied(pos) {
				continue
			}
			g.Obstacles = append(g.Obstacles, pos)
			next := Position{X: x + 1, Y: y}
			if g.rng.Intn(2) == 0 {
				next = Position{X: x, Y: y + 1}
			}
			if next.X < g.Width && next.Y < g.Height && !g.nearSpawn(next) && !g.isPositionOccupied(next) {
				g.Obstacles = append(g.Obstacles, next)
			}
		}
	}
}

// Steer turns the commander, starting the game from the start screen. Turning back
// on yourself is ignored, judged against the last step taken so two quick presses
// within one tick can't reverse into the trail.
func (g *Game) Steer(dir Direction) {
	switch g.State {
	case Ready:
		g.State = Playing
	case Playing:
	default:
		return
	}
	if dir != g.moved.opposite() {
		g.Direction = dir
	}
}

// Start begins a game from the start screen, resumes from pause, or starts over
// after a game over or a win.
func (g *Game) Start() {
	switch g.State {
	case Ready, Paused:
		g.State = Playing
	case GameOver, Won:
		g.reset()
		g.State = Playing
	}
}

// Pause toggles pause while playing.
func (g *Game) Pause() {
	switch g.State {
	case Playing:
		g.State = Paused
	case Paused:
		g.State = Playing
	}
}

// Restart returns to the start screen.
func (g *Game) Restart() { g.reset() }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
