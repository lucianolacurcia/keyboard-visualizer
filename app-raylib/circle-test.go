package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	WINDOW_SIZE = 300
	CIRCLE_RADIUS = 120
)

func main() {
	// Set window configuration flags for overlay behavior
	rl.SetConfigFlags(
		rl.FlagWindowUndecorated |    // No window decorations
		rl.FlagWindowTopmost |        // Always on top
		rl.FlagWindowTransparent,     // Transparent background
		// Removed FlagWindowMousePassthrough to allow interaction
	)

	// Initialize window
	rl.InitWindow(WINDOW_SIZE, WINDOW_SIZE, "Raylib Circle Test")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// Center the window on screen
	screenWidth := rl.GetMonitorWidth(0)
	screenHeight := rl.GetMonitorHeight(0)
	rl.SetWindowPosition(
		(screenWidth-WINDOW_SIZE)/2,
		(screenHeight-WINDOW_SIZE)/2,
	)

	frameCount := 0

	// Main loop
	for !rl.WindowShouldClose() {
		frameCount++

		rl.BeginDrawing()

		// CRITICAL: Clear with fully transparent background
		rl.ClearBackground(rl.NewColor(0, 0, 0, 0)) // Completely transparent

		center := rl.NewVector2(WINDOW_SIZE/2, WINDOW_SIZE/2)

		// Draw animated circle with different opacity layers
		pulseRadius := CIRCLE_RADIUS + 20*float32(math.Sin(float64(frameCount)*0.05))

		// Outer glow (very transparent)
		rl.DrawCircleV(center, pulseRadius+10, rl.NewColor(79, 195, 247, 30))

		// Main circle (semi-transparent)
		rl.DrawCircleV(center, CIRCLE_RADIUS, rl.NewColor(79, 195, 247, 180))

		// Inner circle (more opaque)
		rl.DrawCircleV(center, CIRCLE_RADIUS-30, rl.NewColor(0, 150, 200, 220))

		// Center dot (opaque)
		rl.DrawCircleV(center, 10, rl.White)

		// Text in center
		text := "CIRCLE"
		textSize := int32(20)
		textWidth := rl.MeasureText(text, textSize)
		rl.DrawText(text, (WINDOW_SIZE-textWidth)/2, WINDOW_SIZE/2-10, textSize, rl.Black)

		// Instructions
		infoText := "ESC to close - SPACE for debug info"
		infoWidth := rl.MeasureText(infoText, 12)
		rl.DrawText(infoText, (WINDOW_SIZE-infoWidth)/2, WINDOW_SIZE-30, 12, rl.White)

		// Debug info
		if rl.IsKeyPressed(rl.KeySpace) {
			rl.DrawText("TRANSPARENCY TEST RUNNING!", 20, 20, 16, rl.Red)
		}

		rl.EndDrawing()
	}
}