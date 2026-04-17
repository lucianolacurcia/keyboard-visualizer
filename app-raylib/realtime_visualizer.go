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

	fmt.Printf("✅ Loaded keyboard: %d keys, %d layers\n", len(visualizer.Keys), len(visualizer.Keymap.Layers))

	// Initialize HID reader
	hidReader := NewHIDReader()
	fmt.Print("🔍 Searching for ZMK device...")

	if err := hidReader.FindKeyboardDevice(); err != nil {
		log.Printf("Warning: No ZMK device found: %v", err)
		fmt.Println(" Not found - running in demo mode")
	} else {
		fmt.Println(" Found!")
		if err := hidReader.StartReading(); err != nil {
			log.Fatalf("Failed to start HID reader: %v", err)
		}
		defer hidReader.Stop()
	}

	// Configure window for transparency (from our previous tests)
	rl.SetConfigFlags(
		rl.FlagWindowUndecorated |
		rl.FlagWindowTopmost |
		rl.FlagWindowTransparent,
	)

	// Initialize Raylib window with auto-calculated size
	windowWidth, windowHeight := visualizer.GetWindowSize()
	rl.InitWindow(windowWidth, windowHeight, "Raylib Keyboard Test")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// HID event processing
	eventChan := hidReader.GetEventChannel()
	frameCount := 0

	fmt.Println("🎮 Controls: ESC=quit, 1/2/3=layer switching (if no HID device)")
	fmt.Println("🎯 Real-time mode: Press keys on your Totem to see them light up!")

	// Main loop
	for !rl.WindowShouldClose() {
		frameCount++

		// Process HID events
		for {
			select {
			case hidEvent := <-eventChan:
				processHIDEvent(&hidEvent, visualizer)
			default:
				// No more events, break from processing loop
				goto endEventProcessing
			}
		}
	endEventProcessing:

		// Demo mode if no HID device (for testing)
		if !hidReader.IsConnected() {
			// Keyboard layer switching for testing
			if rl.IsKeyPressed(rl.KeyOne) {
				visualizer.SwitchLayer("BASE")
				fmt.Println("Demo: Switched to BASE layer")
			}
			if rl.IsKeyPressed(rl.KeyTwo) {
				visualizer.SwitchLayer("WM")
				fmt.Println("Demo: Switched to WM layer")
			}
			if rl.IsKeyPressed(rl.KeyThree) {
				visualizer.SwitchLayer("WMS")
				fmt.Println("Demo: Switched to WMS layer")
			}

			// Simple demo animation
			if frameCount%180 == 0 { // Every 3 seconds
				keyIndex := (frameCount / 180) % len(visualizer.Keys)
				if visualizer.Keys[keyIndex].State.IsPressed {
					visualizer.ReleaseKey(keyIndex)
				} else {
					visualizer.PressKey(keyIndex)
				}
			}
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(0, 0, 0, 0)) // Transparent background

		// Draw all keys
		for _, key := range visualizer.Keys {
			// Skip transparent keys
			if key.State.IsTransparent {
				continue
			}

			// Draw key rectangle
			rl.DrawRectangle(
				int32(key.PixelX),
				int32(key.PixelY),
				int32(key.PixelWidth-2), // Small gap between keys
				int32(key.PixelHeight-2),
				key.BackgroundColor,
			)

			// Draw key border
			rl.DrawRectangleLines(
				int32(key.PixelX),
				int32(key.PixelY),
				int32(key.PixelWidth-2),
				int32(key.PixelHeight-2),
				key.BorderColor,
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

			// Draw hold-tap indicator
			if key.KeyEntry.IsHoldTap() {
				holdText := key.KeyEntry.GetHoldBehavior()
				if holdText != "" && holdText != key.KeyEntry.GetDisplayText() {
					smallFontSize := int32(8)
					holdTextWidth := rl.MeasureText(holdText, smallFontSize)
					holdTextX := int32(key.PixelX + key.PixelWidth - float32(holdTextWidth) - 2)
					holdTextY := int32(key.PixelY + 2)
					rl.DrawText(holdText, holdTextX, holdTextY, smallFontSize, rl.Gray)
				}
			}
		}

		// Draw status info
		statusText := fmt.Sprintf("Layer: %s | HID: %s",
			visualizer.CurrentLayer,
			map[bool]string{true: "Connected", false: "Demo Mode"}[hidReader.IsConnected()])
		rl.DrawText(statusText, 10, 10, 16, rl.White)

		if !hidReader.IsConnected() {
			rl.DrawText("Demo: Press 1/2/3 to switch layers", 10, 30, 12, rl.LightGray)
		} else {
			rl.DrawText("Real-time: Press keys on your keyboard!", 10, 30, 12, rl.Green)
		}

		rl.EndDrawing()
	}
}

// processHIDEvent handles incoming HID events
func processHIDEvent(event *HIDEvent, visualizer *KeyboardVisualizer) {
	switch event.Type {
	case REPORT_TYPE_LAYER_STATE:
		if event.LayerState != nil {
			// Handle complete layer state (stateless - no drift!)
			layerName := GetActiveLayerName(event.LayerState.LayerState)
			if layerName != visualizer.CurrentLayer {
				visualizer.SwitchLayer(layerName)
				fmt.Printf("HID: Layer switched to %s (state: 0x%x)\n",
					layerName, event.LayerState.LayerState)
			}
		}

	case REPORT_TYPE_KEY_EVENT:
		if event.KeyEvent != nil {
			// Handle individual key press/release (eventful)
			position := RowColToPosition(event.KeyEvent.Row, event.KeyEvent.Col)
			if position >= 0 && position < len(visualizer.Keys) {
				if event.KeyEvent.Pressed {
					visualizer.PressKey(position)
					fmt.Printf("HID: Key pressed at [%d,%d] -> position %d\n",
						event.KeyEvent.Row, event.KeyEvent.Col, position)
				} else {
					visualizer.ReleaseKey(position)
					fmt.Printf("HID: Key released at [%d,%d] -> position %d\n",
						event.KeyEvent.Row, event.KeyEvent.Col, position)
				}
			}
		}
	}
}