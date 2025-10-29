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

// Position represents a coordinate on the game grid
type Position struct {
	X, Y int
}

// GameState represents the current state of the game
type GameState int

const (
	Playing GameState = iota
	Paused
	GameOver
	LevelComplete
)

// Game represents the main game structure
type Game struct {
	Width, Height    int
	Commander        Position
	Trail            []Position
	Alerts           []Position
	Obstacles        []Position
	Direction        Direction
	State            GameState
	Score            int
	Level            int
	AlertsCollected  int
	AlertsNeeded     int
	StartTime        time.Time
	LastUpdate       time.Time
	LevelCompleteTime time.Time // Time when level was completed
}

// New creates a new game instance
func New(width, height int) *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		Width:     width,
		Height:    height,
		Commander: Position{X: width / 2, Y: height / 2},
		Trail:     make([]Position, 0),
		Alerts:    make([]Position, 0),
		Obstacles: make([]Position, 0),
		Direction: Right,
		State:     Playing,
		Score:     0,
		Level:     1,
		AlertsCollected: 0,
		AlertsNeeded:      5,
		StartTime:         time.Now(),
		LastUpdate:        time.Now(),
		LevelCompleteTime: time.Time{}, // Initialize to zero time
	}
	
	g.spawnAlerts()
	g.setupLevel()
	
	return g
}

// Update updates the game state
func (g *Game) Update() {
	now := time.Now()
	g.LastUpdate = now

	// Always check level completion for timer-based transitions
	g.checkLevelComplete()

	// Only move and check collisions when playing
	if g.State != Playing {
		return
	}

	// Move commander
	g.moveCommander()
	
	// Check collisions
	g.checkCollisions()
}

// moveCommander moves the commander in the current direction
func (g *Game) moveCommander() {
	// Add current position to trail
	g.Trail = append(g.Trail, g.Commander)
	
	// Move commander
	switch g.Direction {
	case Up:
		g.Commander.Y--
	case Down:
		g.Commander.Y++
	case Left:
		g.Commander.X--
	case Right:
		g.Commander.X++
	}
}

// checkCollisions checks for wall, trail, and alert collisions
func (g *Game) checkCollisions() {
	// Wall collision
	if g.Commander.X < 0 || g.Commander.X >= g.Width ||
		g.Commander.Y < 0 || g.Commander.Y >= g.Height {
		g.State = GameOver
		return
	}
	
	// Trail collision (self-collision)
	for _, segment := range g.Trail {
		if g.Commander.X == segment.X && g.Commander.Y == segment.Y {
			g.State = GameOver
			return
		}
	}
	
	// Obstacle collision
	for _, obstacle := range g.Obstacles {
		if g.Commander.X == obstacle.X && g.Commander.Y == obstacle.Y {
			g.State = GameOver
			return
		}
	}
	
	// Alert collision
	for i, alert := range g.Alerts {
		if g.Commander.X == alert.X && g.Commander.Y == alert.Y {
			g.collectAlert(i)
			break
		}
	}
}

// collectAlert handles alert collection
func (g *Game) collectAlert(index int) {
	// Remove the collected alert
	g.Alerts = append(g.Alerts[:index], g.Alerts[index+1:]...)
	
	// Increase score
	basePoints := 10
	comboMultiplier := g.AlertsCollected + 1
	g.Score += basePoints * comboMultiplier
	
	g.AlertsCollected++
	
	// Spawn a new alert
	g.spawnAlerts()
}

// checkLevelComplete checks if the level is complete
func (g *Game) checkLevelComplete() {
	if g.AlertsCollected >= g.AlertsNeeded {
		if g.State != LevelComplete {
			g.State = LevelComplete
			
			// Level completion bonus
			timeBonus := max(0, 60-int(time.Since(g.StartTime).Seconds()))
			g.Score += (100 * g.Level) + timeBonus
			
			// Set a timer to advance to next level after a brief pause
			g.LevelCompleteTime = time.Now()
		} else {
			// Check if enough time has passed (1 second) to advance to next level
			if time.Since(g.LevelCompleteTime) >= 1*time.Second {
				g.nextLevel()
			}
		}
	}
}

// nextLevel advances to the next level
func (g *Game) nextLevel() {
	if g.Level >= 10 {
		// Game completed!
		return
	}
	
	g.Level++
	g.AlertsCollected = 0
	// Progressive difficulty but keep it reasonable
	g.AlertsNeeded = 5 + (g.Level-1) // Level 1: 5, Level 2: 6, ..., Level 10: 14
	g.StartTime = time.Now()
	g.State = Playing
	
	// Reset positions and clear trail for new level
	g.Commander = Position{X: g.Width / 2, Y: g.Height / 2}
	g.Trail = make([]Position, 0) // Reset trail for new level
	
	// Clear alerts and obstacles
	g.Alerts = make([]Position, 0)
	g.Obstacles = make([]Position, 0)
	
	// Setup new level
	g.setupLevel()
	
	// Safety check: Remove any obstacles at commander spawn position
	commanderPos := g.Commander
	for i := len(g.Obstacles) - 1; i >= 0; i-- {
		if g.Obstacles[i].X == commanderPos.X && g.Obstacles[i].Y == commanderPos.Y {
			// Remove obstacle that would collide with commander
			g.Obstacles = append(g.Obstacles[:i], g.Obstacles[i+1:]...)
		}
	}
	
	g.spawnAlerts()
}

// setupLevel configures obstacles and layout for the current level
func (g *Game) setupLevel() {
	switch g.Level {
	case 3, 4: // Static barriers
		g.addStaticBarriers()
	case 5, 6: // Moving obstacles (simplified as static for now)
		g.addStaticBarriers()
		g.addRandomObstacles(2)
	case 7, 8: // More complex layouts
		g.addRandomObstacles(4)
	case 9, 10: // Maximum difficulty
		g.addMazeLayout()
	}
}

// spawnAlerts spawns new alert bubbles
func (g *Game) spawnAlerts() {
	for len(g.Alerts) < 3 { // Keep 3 alerts on screen
		for {
			x := rand.Intn(g.Width)
			y := rand.Intn(g.Height)
			pos := Position{X: x, Y: y}
			
			// Don't spawn on commander, trail, or obstacles
			if !g.isPositionOccupied(pos) {
				g.Alerts = append(g.Alerts, pos)
				break
			}
		}
	}
}

// isPositionOccupied checks if a position is occupied
func (g *Game) isPositionOccupied(pos Position) bool {
	// Check commander
	if g.Commander.X == pos.X && g.Commander.Y == pos.Y {
		return true
	}
	
	// Check trail
	for _, segment := range g.Trail {
		if segment.X == pos.X && segment.Y == pos.Y {
			return true
		}
	}
	
	// Check obstacles
	for _, obstacle := range g.Obstacles {
		if obstacle.X == pos.X && obstacle.Y == pos.Y {
			return true
		}
	}
	
	return false
}

// addStaticBarriers adds static barrier obstacles
func (g *Game) addStaticBarriers() {
	// Add cross pattern with safe distance from center to avoid commander spawn
	centerX, centerY := g.Width/2, g.Height/2
	
	// Create cross pattern with at least 3 cells gap from commander
	// Horizontal barriers (top and bottom of screen)
	for x := 2; x < g.Width-2; x++ {
		// Skip area around commander spawn (leave 3x3 safe zone)
		if abs(x-centerX) > 2 {
			// Top horizontal line
			if centerY-4 >= 0 {
				g.Obstacles = append(g.Obstacles, Position{X: x, Y: centerY - 4})
			}
			// Bottom horizontal line  
			if centerY+4 < g.Height {
				g.Obstacles = append(g.Obstacles, Position{X: x, Y: centerY + 4})
			}
		}
	}
	
	// Vertical barriers (left and right sides)
	for y := 2; y < g.Height-2; y++ {
		// Skip area around commander spawn (leave 3x3 safe zone)
		if abs(y-centerY) > 2 {
			// Left vertical line
			if centerX-4 >= 0 {
				g.Obstacles = append(g.Obstacles, Position{X: centerX - 4, Y: y})
			}
			// Right vertical line
			if centerX+4 < g.Width {
				g.Obstacles = append(g.Obstacles, Position{X: centerX + 4, Y: y})
			}
		}
	}
}

// addRandomObstacles adds random obstacle positions
func (g *Game) addRandomObstacles(count int) {
	centerX, centerY := g.Width/2, g.Height/2
	
	for i := 0; i < count; i++ {
		for attempts := 0; attempts < 50; attempts++ {
			x := rand.Intn(g.Width)
			y := rand.Intn(g.Height)
			pos := Position{X: x, Y: y}
			
			// Don't place obstacles too close to commander spawn (maintain 3x3 safe zone)
			if abs(x-centerX) <= 2 && abs(y-centerY) <= 2 {
				continue
			}
			
			if !g.isPositionOccupied(pos) {
				g.Obstacles = append(g.Obstacles, pos)
				break
			}
		}
	}
}

// addMazeLayout creates a maze-like obstacle layout
func (g *Game) addMazeLayout() {
	// Simple maze pattern that avoids commander spawn area
	centerX, centerY := g.Width/2, g.Height/2
	
	for x := 2; x < g.Width-2; x += 4 {
		for y := 2; y < g.Height-2; y += 4 {
			pos := Position{X: x, Y: y}
			
			// Skip positions too close to commander spawn (maintain 3x3 safe zone)
			if abs(x-centerX) <= 2 && abs(y-centerY) <= 2 {
				continue
			}
			
			if !g.isPositionOccupied(pos) {
				g.Obstacles = append(g.Obstacles, pos)
				
				// Add connecting obstacle
				var nextPos Position
				if rand.Intn(2) == 0 {
					nextPos = Position{X: x + 1, Y: y}
				} else {
					nextPos = Position{X: x, Y: y + 1}
				}
				
				// Check the next position too - maintain safe zone
				if nextPos.X < g.Width && nextPos.Y < g.Height && 
				   !g.isPositionOccupied(nextPos) &&
				   !(abs(nextPos.X-centerX) <= 2 && abs(nextPos.Y-centerY) <= 2) {
					g.Obstacles = append(g.Obstacles, nextPos)
				}
			}
		}
	}
}

// Public getters
func (g *Game) GetCommander() Position { return g.Commander }
func (g *Game) GetTrail() []Position { return g.Trail }
func (g *Game) GetAlerts() []Position { return g.Alerts }
func (g *Game) GetObstacles() []Position { return g.Obstacles }
func (g *Game) GetScore() int { return g.Score }
func (g *Game) GetLevel() int { return g.Level }
func (g *Game) GetAlertsCollected() int { return g.AlertsCollected }
func (g *Game) GetAlertsNeeded() int { return g.AlertsNeeded }
func (g *Game) GetState() GameState { return g.State }
func (g *Game) GetWidth() int { return g.Width }
func (g *Game) GetHeight() int { return g.Height }
func (g *Game) IsRunning() bool { return g.State == Playing }

// Control methods
func (g *Game) SetDirection(dir Direction) {
	// Prevent immediate reversal
	opposite := map[Direction]Direction{
		Up: Down, Down: Up, Left: Right, Right: Left,
	}
	if g.Direction != opposite[dir] {
		g.Direction = dir
	}
}

func (g *Game) Pause() {
	if g.State == Playing {
		g.State = Paused
	} else if g.State == Paused {
		g.State = Playing
	}
}

func (g *Game) Restart() {
	*g = *New(g.Width, g.Height)
}

// Utility functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}