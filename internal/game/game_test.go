package game

import (
	"math/rand"
	"testing"
	"time"
)

// testGame returns a seeded game with a controllable clock.
func testGame(t *testing.T) (*Game, *time.Time) {
	t.Helper()
	clock := time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC)
	g := newGame(20, 20, rand.New(rand.NewSource(1)), func() time.Time { return clock })
	return g, &clock
}

func TestStartsOnTheStartScreen(t *testing.T) {
	g, _ := testGame(t)
	if g.State != Ready {
		t.Fatalf("state = %v, want Ready", g.State)
	}
	g.Update()
	if g.Commander != (Position{10, 10}) {
		t.Fatalf("commander moved before the game started: %v", g.Commander)
	}
	g.Steer(Up)
	if g.State != Playing || g.Direction != Up {
		t.Fatalf("steering should start the game heading up; state=%v dir=%v", g.State, g.Direction)
	}
}

func TestCannotReverseWithTwoQuickTurns(t *testing.T) {
	g, _ := testGame(t)
	g.Start()
	g.Alerts = nil // keep the path clear
	g.Update()     // moving right
	g.Steer(Up)    // two presses within one tick...
	g.Steer(Left)  // ...must not turn straight back into the trail
	if g.Direction != Up {
		t.Fatalf("direction = %v, want Up (Left is a reversal of the last step)", g.Direction)
	}
	g.Update()
	if g.State != Playing {
		t.Fatalf("state = %v after a legal turn, want Playing", g.State)
	}
}

func TestWallEndsTheGame(t *testing.T) {
	g, _ := testGame(t)
	g.Start()
	g.Alerts = nil
	for i := 0; i < 11 && g.State == Playing; i++ {
		g.Alerts = nil
		g.Update()
	}
	if g.State != GameOver {
		t.Fatalf("state = %v after running into the right wall, want GameOver", g.State)
	}
	if g.Commander != (Position{19, 10}) {
		t.Fatalf("commander = %v, want it left at the wall (19,10)", g.Commander)
	}
	if ev := g.Events(); len(ev) != 1 || ev[0] != EventGameOver {
		t.Fatalf("events = %v, want [EventGameOver]", ev)
	}
	g.Start()
	if g.State != Playing || g.Level != 1 || g.Score != 0 {
		t.Fatalf("start after game over should begin a fresh game; state=%v level=%d score=%d", g.State, g.Level, g.Score)
	}
}

// collectNext puts an alert directly ahead and steps onto it.
func collectNext(g *Game) {
	g.Commander = Position{1, 1}
	g.Trail = nil
	g.Obstacles = nil
	g.Direction, g.moved = Right, Right
	g.Alerts = []Position{{2, 1}}
	g.Update()
}

func TestScoringAndLevelUp(t *testing.T) {
	g, clock := testGame(t)
	g.Start()
	for i := 1; i < g.AlertsNeeded; i++ {
		collectNext(g)
	}
	if g.Score != 10+20+30+40 {
		t.Fatalf("score after 4 alerts = %d, want 100", g.Score)
	}
	*clock = clock.Add(20 * time.Second)
	collectNext(g) // fifth alert clears level 1
	if g.State != LevelComplete {
		t.Fatalf("state = %v, want LevelComplete", g.State)
	}
	if want := 150 + 100 + 40; g.Score != want {
		t.Fatalf("score = %d, want %d (alerts + level bonus + time bonus)", g.Score, want)
	}
	g.Update()
	if g.Level != 1 {
		t.Fatal("advanced before the level-clear pause ended")
	}
	*clock = clock.Add(levelPause)
	g.Update()
	if g.Level != 2 || g.State != Playing || g.AlertsNeeded != 6 || len(g.Trail) != 0 {
		t.Fatalf("level 2 not set up: level=%d state=%v needed=%d trail=%d", g.Level, g.State, g.AlertsNeeded, len(g.Trail))
	}
}

func TestClearingLevelTenWins(t *testing.T) {
	g, _ := testGame(t)
	g.Start()
	g.Level = MaxLevel
	g.startLevel()
	for i := 0; i < g.AlertsNeeded; i++ {
		collectNext(g)
	}
	if g.State != Won {
		t.Fatalf("state = %v after clearing level 10, want Won", g.State)
	}
}

func TestLevelsKeepTheSpawnClear(t *testing.T) {
	g, _ := testGame(t)
	for level := 1; level <= MaxLevel; level++ {
		g.Level = level
		g.startLevel()
		for _, o := range g.Obstacles {
			if g.nearSpawn(o) {
				t.Fatalf("level %d: obstacle %v inside the spawn safe zone", level, o)
			}
		}
		if len(g.Alerts) != 3 {
			t.Fatalf("level %d: %d alerts, want 3", level, len(g.Alerts))
		}
	}
}

func TestTickGetsFaster(t *testing.T) {
	if TickInterval(1) <= TickInterval(5) || TickInterval(10) != 125*time.Millisecond {
		t.Fatalf("tick intervals: L1=%v L5=%v L10=%v", TickInterval(1), TickInterval(5), TickInterval(10))
	}
}
