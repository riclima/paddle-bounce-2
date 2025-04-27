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
	Over        bool
	Paused      bool
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
	}

	// Randomly choose which player to target initially
	targetPlayer := PlayerID(Left)
	if rand.Float32() > 0.5 {
		targetPlayer = PlayerID(Right)
	}

	game.ResetBall(targetPlayer)

	return game
}

func (g *Game) Update() {
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

	if g.Ball.Position.X-g.Ball.Radius <= g.PaddleLeft.Position.X+g.PaddleLeft.Dimensions.X &&
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

	// Right paddle
	if g.Ball.Position.X+g.Ball.Radius >= g.PaddleRight.Position.X &&
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

	// Check for scoring
	if g.Ball.Position.X < 0 {
		g.ScoreBoard.Right++
		g.ResetBall(Right)
	} else if g.Ball.Position.X > float32(WindowWidth) {
		g.ScoreBoard.Left++
		g.ResetBall(Left)
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

func UpdateDrawFrame(game *Game) {
	game.Update()

	rl.BeginDrawing()
	defer rl.EndDrawing()

	rl.ClearBackground(rl.RayWhite)

	renderers := []Renderer{
		&game.Net,
		&game.PaddleLeft,
		&game.PaddleRight,
		&game.Ball,
		&game.ScoreBoard,
	}

	for _, renderer := range renderers {
		renderer.Render()
	}

	rl.DrawFPS(WindowWidth-100, 10)
}
