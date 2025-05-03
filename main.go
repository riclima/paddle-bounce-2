package main

import (
	"fmt"
	"github.com/charmbracelet/log"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
)

const WindowWidth = 1024
const WindowHeight = 768
const MaxBounceAngle = 75.0 * math.Pi / 180.0 // 75 degrees in radians
const WinScore = 21                           // Score needed to win the game

type GameState int

const (
	TitleScreen GameState = iota
	GameplayScreen
	PauseScreen
	EndGameScreen
)

type Renderer interface {
	Render()
}

type PlayerID uint8

const (
	Left = iota
	Right
)

type Paddle struct {
	ID         PlayerID
	Position   rl.Vector2
	Dimensions rl.Vector2
}

func (p *Paddle) Render() {
	rl.DrawRectangleV(p.Position, p.Dimensions, rl.Black)
}

type Ball struct {
	Position rl.Vector2
	Velocity rl.Vector2
	Radius   float32
}

func (b *Ball) Render() {
	size := b.Radius * 2
	rl.DrawRectangle(
		int32(b.Position.X-b.Radius),
		int32(b.Position.Y-b.Radius),
		int32(size),
		int32(size),
		rl.Black,
	)
}

type Net struct{}

func (n *Net) Render() {
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()
	centerX := screenWidth / 2
	dashWidth := 4

	dashCount := 24
	dashUnits := 3
	spaceUnits := 1
	unitTotal := (dashCount * dashUnits) + ((dashCount - 1) * spaceUnits)
	unitHeight := screenHeight / unitTotal

	dashHeight := unitHeight * dashUnits
	spaceHeight := unitHeight * spaceUnits

	for i := 0; i < 24; i++ {
		y := i * (dashHeight + spaceHeight)
		rl.DrawRectangle(
			int32(centerX-dashWidth/2),
			int32(y),
			int32(dashWidth),
			int32(dashHeight),
			rl.Black,
		)
	}
}

type ScoreBoard struct {
	Left  uint
	Right uint
}

func (s *ScoreBoard) Render() {
	leftScore := fmt.Sprintf("%d", s.Left)
	rightScore := fmt.Sprintf("%d", s.Right)

	fontSize := 60

	screenWidth := rl.GetScreenWidth()
	leftX := screenWidth/4 - (len(leftScore)*fontSize)/4
	rightX := 3*screenWidth/4 - (len(rightScore)*fontSize)/4

	y := 50
	rl.DrawText(leftScore, int32(leftX), int32(y), int32(fontSize), rl.Black)
	rl.DrawText(rightScore, int32(rightX), int32(y), int32(fontSize), rl.Black)
}

type Game struct {
	PaddleLeft  Paddle
	PaddleRight Paddle
	Ball        Ball
	Net         Net
	ScoreBoard  ScoreBoard
	State       GameState
	Winner      PlayerID
}

func (g *Game) ResetBall(targetPlayer PlayerID) {
	g.Ball.Position = rl.Vector2{
		X: float32(WindowWidth) / 2,
		Y: float32(WindowHeight) / 2,
	}

	ballSpeed := float32(4.0)

	directionX := float32(1.0)
	if targetPlayer == Left {
		directionX = -1.0
	}

	distanceToGoal := float32(WindowWidth / 2)

	maxYDisplacement := float32(WindowHeight / 2)

	randomYDisplacement := (rand.Float32()*2 - 1) * maxYDisplacement

	velocityY := (randomYDisplacement / distanceToGoal) * ballSpeed

	g.Ball.Velocity = rl.Vector2{
		X: directionX * ballSpeed,
		Y: velocityY,
	}
}

func NewGame() Game {
	game := Game{
		PaddleLeft: Paddle{
			ID:         Left,
			Position:   rl.Vector2{X: 50, Y: (float32(rl.GetScreenHeight()) / 2) - 50},
			Dimensions: rl.Vector2{X: 20, Y: 100},
		},
		PaddleRight: Paddle{
			ID:         Right,
			Position:   rl.Vector2{X: float32(rl.GetScreenWidth()) - 70, Y: (float32(rl.GetScreenHeight()) / 2) - 50},
			Dimensions: rl.Vector2{X: 20, Y: 100},
		},
		Ball: Ball{
			Position: rl.Vector2{X: float32(rl.GetScreenWidth()) / 2, Y: float32(rl.GetScreenHeight()) / 2},
			Velocity: rl.Vector2{X: 0, Y: 0}, // Will be set by ResetBall
			Radius:   8,
		},
		Net: Net{},
		ScoreBoard: ScoreBoard{
			Left:  0,
			Right: 0,
		},
		State: TitleScreen,
	}

	targetPlayer := PlayerID(Left)
	if rand.Float32() > 0.5 {
		targetPlayer = PlayerID(Right)
	}

	game.ResetBall(targetPlayer)

	return game
}

func (g *Game) Update() {
	// Handle state transitions based on key presses
	if rl.IsKeyPressed(rl.KeyP) && g.State == GameplayScreen {
		g.State = PauseScreen
		return
	} else if rl.IsKeyPressed(rl.KeyP) && g.State == PauseScreen {
		g.State = GameplayScreen
		return
	}

	switch g.State {
	case TitleScreen:
		if rl.IsKeyPressed(rl.KeyEnter) {
			g.State = GameplayScreen
		}
		return

	case PauseScreen:
		return

	case EndGameScreen:
		if rl.IsKeyPressed(rl.KeyEnter) {
			*g = NewGame()
			g.State = GameplayScreen
		}
		return

	case GameplayScreen:
		paddleSpeed := float32(600.0)
		deltaTime := rl.GetFrameTime()
		frameSpeed := paddleSpeed * deltaTime

		if rl.IsKeyDown(rl.KeyW) {
			g.PaddleLeft.Position.Y -= frameSpeed
		}
		if rl.IsKeyDown(rl.KeyS) {
			g.PaddleLeft.Position.Y += frameSpeed
		}

		if rl.IsKeyDown(rl.KeyUp) {
			g.PaddleRight.Position.Y -= frameSpeed
		}
		if rl.IsKeyDown(rl.KeyDown) {
			g.PaddleRight.Position.Y += frameSpeed
		}

		if g.PaddleLeft.Position.Y < 0 {
			g.PaddleLeft.Position.Y = 0
		}
		if g.PaddleLeft.Position.Y+g.PaddleLeft.Dimensions.Y > float32(WindowHeight) {
			g.PaddleLeft.Position.Y = float32(WindowHeight) - g.PaddleLeft.Dimensions.Y
		}
		if g.PaddleRight.Position.Y < 0 {
			g.PaddleRight.Position.Y = 0
		}
		if g.PaddleRight.Position.Y+g.PaddleRight.Dimensions.Y > float32(WindowHeight) {
			g.PaddleRight.Position.Y = float32(WindowHeight) - g.PaddleRight.Dimensions.Y
		}

		g.Ball.Position.X += g.Ball.Velocity.X
		g.Ball.Position.Y += g.Ball.Velocity.Y

		if g.Ball.Position.Y-g.Ball.Radius <= 0 || g.Ball.Position.Y+g.Ball.Radius >= float32(WindowHeight) {
			g.Ball.Velocity.Y = -g.Ball.Velocity.Y
		}

		// Left paddle front collision
		if g.Ball.Position.X-g.Ball.Radius <= g.PaddleLeft.Position.X+g.PaddleLeft.Dimensions.X &&
			g.Ball.Position.X-g.Ball.Radius >= g.PaddleLeft.Position.X &&
			g.Ball.Position.Y >= g.PaddleLeft.Position.Y &&
			g.Ball.Position.Y <= g.PaddleLeft.Position.Y+g.PaddleLeft.Dimensions.Y &&
			g.Ball.Velocity.X < 0 {

			paddleCenter := g.PaddleLeft.Position.Y + (g.PaddleLeft.Dimensions.Y / 2)
			relativeIntersectY := paddleCenter - g.Ball.Position.Y

			normalizedRelativeIntersectionY := relativeIntersectY / (g.PaddleLeft.Dimensions.Y / 2)

			bounceAngle := normalizedRelativeIntersectionY * MaxBounceAngle

			ballSpeed := float32(math.Sqrt(float64(g.Ball.Velocity.X*g.Ball.Velocity.X + g.Ball.Velocity.Y*g.Ball.Velocity.Y)))

			g.Ball.Velocity.X = ballSpeed * float32(math.Cos(float64(bounceAngle)))
			g.Ball.Velocity.Y = ballSpeed * float32(-math.Sin(float64(bounceAngle)))
		}

		// Left paddle top/bottom collision
		if g.Ball.Position.X >= g.PaddleLeft.Position.X &&
			g.Ball.Position.X <= g.PaddleLeft.Position.X+g.PaddleLeft.Dimensions.X {

			// Top edge collision
			if g.Ball.Position.Y+g.Ball.Radius >= g.PaddleLeft.Position.Y &&
				g.Ball.Position.Y-g.Ball.Radius <= g.PaddleLeft.Position.Y &&
				g.Ball.Velocity.Y > 0 {
				g.Ball.Velocity.Y = -g.Ball.Velocity.Y
			}

			// Bottom edge collision
			if g.Ball.Position.Y-g.Ball.Radius <= g.PaddleLeft.Position.Y+g.PaddleLeft.Dimensions.Y &&
				g.Ball.Position.Y+g.Ball.Radius >= g.PaddleLeft.Position.Y+g.PaddleLeft.Dimensions.Y &&
				g.Ball.Velocity.Y < 0 {
				g.Ball.Velocity.Y = -g.Ball.Velocity.Y
			}
		}

		// Right paddle front collision
		if g.Ball.Position.X+g.Ball.Radius >= g.PaddleRight.Position.X &&
			g.Ball.Position.X+g.Ball.Radius <= g.PaddleRight.Position.X+g.PaddleRight.Dimensions.X &&
			g.Ball.Position.Y >= g.PaddleRight.Position.Y &&
			g.Ball.Position.Y <= g.PaddleRight.Position.Y+g.PaddleRight.Dimensions.Y &&
			g.Ball.Velocity.X > 0 {

			paddleCenter := g.PaddleRight.Position.Y + (g.PaddleRight.Dimensions.Y / 2)
			relativeIntersectY := paddleCenter - g.Ball.Position.Y

			normalizedRelativeIntersectionY := relativeIntersectY / (g.PaddleRight.Dimensions.Y / 2)

			bounceAngle := normalizedRelativeIntersectionY * MaxBounceAngle

			ballSpeed := float32(math.Sqrt(float64(g.Ball.Velocity.X*g.Ball.Velocity.X + g.Ball.Velocity.Y*g.Ball.Velocity.Y)))

			g.Ball.Velocity.X = -ballSpeed * float32(math.Cos(float64(bounceAngle)))
			g.Ball.Velocity.Y = ballSpeed * float32(-math.Sin(float64(bounceAngle)))
		}

		// Right paddle top/bottom collision
		if g.Ball.Position.X >= g.PaddleRight.Position.X &&
			g.Ball.Position.X <= g.PaddleRight.Position.X+g.PaddleRight.Dimensions.X {

			// Top edge collision
			if g.Ball.Position.Y+g.Ball.Radius >= g.PaddleRight.Position.Y &&
				g.Ball.Position.Y-g.Ball.Radius <= g.PaddleRight.Position.Y &&
				g.Ball.Velocity.Y > 0 {
				g.Ball.Velocity.Y = -g.Ball.Velocity.Y
			}

			// Bottom edge collision
			if g.Ball.Position.Y-g.Ball.Radius <= g.PaddleRight.Position.Y+g.PaddleRight.Dimensions.Y &&
				g.Ball.Position.Y+g.Ball.Radius >= g.PaddleRight.Position.Y+g.PaddleRight.Dimensions.Y &&
				g.Ball.Velocity.Y < 0 {
				g.Ball.Velocity.Y = -g.Ball.Velocity.Y
			}
		}

		// Check for scoring
		if g.Ball.Position.X < 0 {
			g.ScoreBoard.Right++
			g.ResetBall(Left) // Serve to the opponent of the player who scored

			// Check for win condition
			if g.ScoreBoard.Right >= WinScore {
				g.Winner = Right
				g.State = EndGameScreen
			}
		} else if g.Ball.Position.X > float32(WindowWidth) {
			g.ScoreBoard.Left++
			g.ResetBall(Right) // Serve to the opponent of the player who scored

			// Check for win condition
			if g.ScoreBoard.Left >= WinScore {
				g.Winner = Left
				g.State = EndGameScreen
			}
		}
	}
}

func main() {
	welcomeMessage := "Get ready to BOUNCE"
	log.Info(welcomeMessage)

	rl.SetConfigFlags(rl.FlagVsyncHint)
	rl.InitWindow(WindowWidth, WindowHeight, "Paddle Bounce 2")
	defer rl.CloseWindow()

	game := NewGame()
	for !rl.WindowShouldClose() {
		UpdateDrawFrame(&game)
	}
}

func (g *Game) RenderTitleScreen() {
	rl.ClearBackground(rl.RayWhite)

	title := "PADDLE BOUNCE 2"
	subtitle := "THE REBOUND"
	instructions := "PRESS ENTER TO START"
	controls := "CONTROLS:"
	leftControls := "PLAYER 1: W/S KEYS"
	rightControls := "PLAYER 2: UP/DOWN KEYS"
	pauseControl := "PAUSE: P KEY"

	titleFontSize := 60
	subtitleFontSize := 30
	instructionsFontSize := 40
	controlsFontSize := 20

	titleWidth := rl.MeasureText(title, int32(titleFontSize))
	subtitleWidth := rl.MeasureText(subtitle, int32(subtitleFontSize))
	instructionsWidth := rl.MeasureText(instructions, int32(instructionsFontSize))

	titleX := (WindowWidth - int(titleWidth)) / 2
	subtitleX := (WindowWidth - int(subtitleWidth)) / 2
	instructionsX := (WindowWidth - int(instructionsWidth)) / 2

	rl.DrawText(title, int32(titleX), 150, int32(titleFontSize), rl.Black)
	rl.DrawText(subtitle, int32(subtitleX), 220, int32(subtitleFontSize), rl.DarkGray)
	rl.DrawText(instructions, int32(instructionsX), 350, int32(instructionsFontSize), rl.Black)

	controlsX := WindowWidth/2 - 100
	rl.DrawText(controls, int32(controlsX), 450, int32(controlsFontSize), rl.DarkGray)
	rl.DrawText(leftControls, int32(controlsX), 480, int32(controlsFontSize), rl.DarkGray)
	rl.DrawText(rightControls, int32(controlsX), 510, int32(controlsFontSize), rl.DarkGray)
	rl.DrawText(pauseControl, int32(controlsX), 540, int32(controlsFontSize), rl.DarkGray)

	// Draw a bouncing ball animation
	time := float32(rl.GetTime())
	ballX := WindowWidth/2 + int(math.Sin(float64(time*2))*200)
	ballY := 300 + int(math.Cos(float64(time*2))*50)
	rl.DrawRectangle(int32(ballX-8), int32(ballY-8), 16, 16, rl.Black)
}

func (g *Game) RenderPauseScreen() {
	// First, render the game screen in the background
	renderers := []Renderer{
		&g.Net,
		&g.PaddleLeft,
		&g.PaddleRight,
		&g.Ball,
		&g.ScoreBoard,
	}

	for _, renderer := range renderers {
		renderer.Render()
	}

	// Draw a semi-transparent overlay
	rl.DrawRectangle(0, 0, int32(WindowWidth), int32(WindowHeight), rl.ColorAlpha(rl.Black, 0.5))

	// Draw pause text
	pauseText := "GAME PAUSED"
	instructions := "PRESS P TO RESUME"

	pauseFontSize := 60
	instructionsFontSize := 30

	pauseWidth := rl.MeasureText(pauseText, int32(pauseFontSize))
	instructionsWidth := rl.MeasureText(instructions, int32(instructionsFontSize))

	pauseX := (WindowWidth - int(pauseWidth)) / 2
	instructionsX := (WindowWidth - int(instructionsWidth)) / 2

	rl.DrawText(pauseText, int32(pauseX), 300, int32(pauseFontSize), rl.White)
	rl.DrawText(instructions, int32(instructionsX), 380, int32(instructionsFontSize), rl.White)
}

func (g *Game) RenderEndGameScreen() {
	rl.ClearBackground(rl.RayWhite)

	var winnerText string
	if g.Winner == Left {
		winnerText = "PLAYER 1 WINS!"
	} else {
		winnerText = "PLAYER 2 WINS!"
	}

	gameOverText := "GAME OVER"
	scoreText := fmt.Sprintf("FINAL SCORE: %d - %d", g.ScoreBoard.Left, g.ScoreBoard.Right)
	instructions := "PRESS ENTER TO PLAY AGAIN"

	gameOverFontSize := 60
	winnerFontSize := 50
	scoreFontSize := 30
	instructionsFontSize := 30

	gameOverWidth := rl.MeasureText(gameOverText, int32(gameOverFontSize))
	winnerWidth := rl.MeasureText(winnerText, int32(winnerFontSize))
	scoreWidth := rl.MeasureText(scoreText, int32(scoreFontSize))
	instructionsWidth := rl.MeasureText(instructions, int32(instructionsFontSize))

	gameOverX := (WindowWidth - int(gameOverWidth)) / 2
	winnerX := (WindowWidth - int(winnerWidth)) / 2
	scoreX := (WindowWidth - int(scoreWidth)) / 2
	instructionsX := (WindowWidth - int(instructionsWidth)) / 2

	rl.DrawText(gameOverText, int32(gameOverX), 200, int32(gameOverFontSize), rl.Black)
	rl.DrawText(winnerText, int32(winnerX), 280, int32(winnerFontSize), rl.Black)
	rl.DrawText(scoreText, int32(scoreX), 350, int32(scoreFontSize), rl.DarkGray)
	rl.DrawText(instructions, int32(instructionsX), 450, int32(instructionsFontSize), rl.Black)
}

func (g *Game) RenderGameplayScreen() {
	rl.ClearBackground(rl.RayWhite)

	renderers := []Renderer{
		&g.Net,
		&g.PaddleLeft,
		&g.PaddleRight,
		&g.Ball,
		&g.ScoreBoard,
	}

	for _, renderer := range renderers {
		renderer.Render()
	}
}

func UpdateDrawFrame(game *Game) {
	game.Update()

	rl.BeginDrawing()
	defer rl.EndDrawing()

	// Render the appropriate screen based on the game state
	switch game.State {
	case TitleScreen:
		game.RenderTitleScreen()
	case GameplayScreen:
		game.RenderGameplayScreen()
	case PauseScreen:
		game.RenderPauseScreen()
	case EndGameScreen:
		game.RenderEndGameScreen()
	}

	rl.DrawFPS(WindowWidth-100, 10)
}
