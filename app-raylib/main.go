package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	WINDOW_WIDTH  = 400
	WINDOW_HEIGHT = 200
)

func main() {
	// Set window configuration flags for overlay behavior
	rl.SetConfigFlags(
		rl.FlagWindowUndecorated |    // No window decorations
		rl.FlagWindowTopmost |        // Always on top
		rl.FlagWindowTransparent |    // Transparent background
		rl.FlagWindowMousePassthrough, // Click-through (optional)
	)

	// Initialize window
	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "Raylib Transparency Test")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// Center the window on screen
	screenWidth := rl.GetMonitorWidth(0)
	screenHeight := rl.GetMonitorHeight(0)
	rl.SetWindowPosition(
		(screenWidth-WINDOW_WIDTH)/2,
		(screenHeight-WINDOW_HEIGHT)/2,
	)

	frameCount := 0

	// Main loop
	for !rl.WindowShouldClose() {
		frameCount++

		rl.BeginDrawing()

		// Clear with fully transparent background
		rl.ClearBackground(rl.NewColor(0, 0, 0, 0)) // RGBA(0,0,0,0) = transparent

		// Draw test elements to verify overlay is working

		// Semi-transparent background for text readability
		rl.DrawRectangle(10, 10, 380, 180, rl.NewColor(0, 0, 0, 128))

		// Test text
		rl.DrawText("🔧 RAYLIB TRANSPARENCY TEST", 20, 30, 20, rl.White)
		rl.DrawText("If you see this floating over other windows,", 20, 60, 16, rl.LightGray)
		rl.DrawText("the overlay is working!", 20, 80, 16, rl.LightGray)

		// Animated element to show it's running
		pulseAlpha := uint8(128 + 127*math.Sin(float64(frameCount)*0.1))
		rl.DrawRectangle(20, 110, 360, 30, rl.NewColor(79, 195, 247, pulseAlpha))
		rl.DrawText("Animated element (proves it's running)", 30, 120, 14, rl.Black)

		// Window info
		rl.DrawText("Press ESC to close", 20, 150, 12, rl.Yellow)
		rl.DrawText("Test: Transparent + Always On Top", 20, 170, 12, rl.Green)

		rl.EndDrawing()
	}
}