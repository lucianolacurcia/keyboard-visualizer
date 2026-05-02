package main

import (
	"fmt"
	"log"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Parse command line arguments - FAIL-FAST approach
	if len(os.Args) < 3 {
		log.Fatal("Usage: keyboard-visualizer <layout.json> <keymap.yaml>\n" +
			"  layout.json: Physical keyboard layout (from keymap-drawer)\n" +
			"  keymap.yaml: User keymap configuration (from keymap-drawer)")
	}

	layoutFile := os.Args[1] // Physical layout (from keymap-drawer)
	keymapFile := os.Args[2] // User keymap (from keymap-drawer)

	// Load keyboard data
	visualizer, err := NewKeyboardVisualizer(keymapFile, layoutFile)
	if err != nil {
		log.Fatalf("FATAL: Error creating visualizer: %v", err)
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

	// Enable mouse passthrough for true clickthrough overlay
	rl.SetWindowState(rl.FlagWindowMousePassthrough)

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

		// Draw combo connections
		drawComboConnections(visualizer)

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
				log.Printf("LAYER:    %s (state: 0x%x)",
					layerName, event.LayerState.LayerState)
			}
		}

	case REPORT_TYPE_KEY_EVENT:
		if event.KeyEvent != nil {
			// Handle individual key press/release (eventful)
			position := int(event.KeyEvent.Position) // Direct position from generic firmware

			// STRICT validation - fail-fast on invalid position
			if position < 0 || position >= len(visualizer.Keys) {
				log.Printf("FATAL: Firmware sent invalid position %d (max: %d)",
					position, len(visualizer.Keys)-1)
				break // Skip this invalid event
			}

			if event.KeyEvent.Pressed {
				visualizer.PressKey(position)
				log.Printf("KEY:      pressed pos %d", position)
			} else {
				visualizer.ReleaseKey(position)
				log.Printf("KEY:      released pos %d", position)
			}
		}
	}
}

// drawComboConnections renders visual indicators for available combos
func drawComboConnections(visualizer *KeyboardVisualizer) {
	// Only show combos for current layer
	currentLayer := visualizer.CurrentLayer

	for _, combo := range visualizer.Keymap.Combos {
		// Check if combo is active in current layer
		isActiveInLayer := len(combo.L) == 0 // No layer restriction = active in all layers
		for _, layer := range combo.L {
			if layer == currentLayer {
				isActiveInLayer = true
				break
			}
		}

		if !isActiveInLayer {
			continue
		}

		// Draw connections between combo keys
		if len(combo.P) >= 2 {
			// Get positions of combo keys
			var keyPositions []rl.Vector2

			for _, pos := range combo.P {
				if pos < len(visualizer.Keys) {
					key := &visualizer.Keys[pos]
					centerX := key.PixelX + key.PixelWidth/2
					centerY := key.PixelY + key.PixelHeight/2
					keyPositions = append(keyPositions, rl.NewVector2(centerX, centerY))
				}
			}

			// Draw lines connecting combo keys
			for i := 0; i < len(keyPositions)-1; i++ {
				rl.DrawLineV(keyPositions[i], keyPositions[i+1], rl.Orange)

				// Draw small circles at connection points
				rl.DrawCircleV(keyPositions[i], 3, rl.Orange)
			}
			if len(keyPositions) > 1 {
				rl.DrawCircleV(keyPositions[len(keyPositions)-1], 3, rl.Orange)
			}

			// Draw combo result text near the connection
			if len(keyPositions) >= 2 {
				// Position text near the center of the combo
				textX := (keyPositions[0].X + keyPositions[len(keyPositions)-1].X) / 2
				textY := (keyPositions[0].Y+keyPositions[len(keyPositions)-1].Y)/2 - 20

				comboText := combo.K.GetDisplayText()
				if comboText != "" {
					// Draw background for text
					textWidth := rl.MeasureText(comboText, 12)
					rl.DrawRectangle(int32(textX-float32(textWidth)/2-2), int32(textY-2),
						int32(textWidth+4), 16, rl.NewColor(0, 0, 0, 128))

					// Draw text
					rl.DrawText(comboText, int32(textX-float32(textWidth)/2), int32(textY),
						12, rl.Orange)
				}
			}
		}
	}
}
