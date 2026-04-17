package main

import (
	"fmt"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Load keyboard data
	visualizer, err := NewKeyboardVisualizer("totem.yaml", "totem.json")
	if err != nil {
		log.Fatalf("Error creating visualizer: %v", err)
	}

	fmt.Printf("✅ Loaded %d keys\n", len(visualizer.Keys))
	fmt.Printf("📐 Window size: %dx%d\n", visualizer.Config.WindowWidth, visualizer.Config.WindowHeight)

	// Initialize Raylib window
	rl.InitWindow(visualizer.Config.WindowWidth, visualizer.Config.WindowHeight, "Raylib Keyboard Test")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	frameCount := 0

	// Main loop
	for !rl.WindowShouldClose() {
		frameCount++

		// Simple key press simulation (for testing)
		if frameCount%120 == 0 { // Every 2 seconds
			keyIndex := (frameCount / 120) % len(visualizer.Keys)
			if visualizer.Keys[keyIndex].State.IsPressed {
				visualizer.ReleaseKey(keyIndex)
			} else {
				visualizer.PressKey(keyIndex)
			}
		}

		// Layer switching test
		if rl.IsKeyPressed(rl.KeyOne) {
			visualizer.SwitchLayer("BASE")
			fmt.Println("Switched to BASE layer")
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			visualizer.SwitchLayer("WM")
			fmt.Println("Switched to WM layer")
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			visualizer.SwitchLayer("WMS")
			fmt.Println("Switched to WMS layer")
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(0, 0, 0, 0)) // Transparent background

		// Draw all keys as simple rectangles
		for _, key := range visualizer.Keys {
			// Skip transparent keys for now
			if key.State.IsTransparent {
				continue
			}

			// Simple rectangle (no rotation)
			rl.DrawRectangle(
				int32(key.PixelX),
				int32(key.PixelY),
				int32(key.PixelWidth-2), // Small gap between keys
				int32(key.PixelHeight-2),
				key.BackgroundColor,
			)

			// Draw key text
			text := key.KeyEntry.GetDisplayText()
			if text != "" {
				textSize := key.FontSize
				textWidth := rl.MeasureText(text, textSize)

				// Center text on key
				textX := int32(key.PixelX + key.PixelWidth/2 - float32(textWidth)/2)
				textY := int32(key.PixelY + key.PixelHeight/2 - float32(textSize)/2)

				rl.DrawText(text, textX, textY, textSize, key.TextColor)
			}
		}

		// Draw info
		infoText := fmt.Sprintf("Layer: %s | Keys: %d | 1/2/3 to switch layers",
			visualizer.CurrentLayer, len(visualizer.Keys))
		rl.DrawText(infoText, 10, 10, 16, rl.White)

		// Draw controls
		rl.DrawText("Press 1=BASE, 2=WM, 3=WMS | ESC=quit", 10, 30, 12, rl.LightGray)

		rl.EndDrawing()
	}
}