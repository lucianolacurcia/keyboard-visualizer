package main

import (
	"fmt"
	"log"
)

func main() {
	// Load the keymap
	keymap, err := LoadKeymap("totem.yaml")
	if err != nil {
		log.Fatalf("Error loading keymap: %v", err)
	}

	fmt.Printf("🎹 Loaded keymap for: %s\n\n", keymap.Layout.ZmkKeyboard)
	fmt.Printf("📊 Stats:\n")
	fmt.Printf("  - Layers: %d\n", len(keymap.Layers))
	fmt.Printf("  - Keys per layer: %d\n", keymap.GetKeyCount())
	fmt.Printf("  - Combos: %d\n\n", len(keymap.Combos))

	// Show layer names
	fmt.Printf("🗂️  Layers: %v\n\n", keymap.GetLayerNames())

	// Show first 10 keys of BASE layer
	fmt.Printf("🔤 BASE layer (first 10 keys):\n")
	for i := 0; i < 10; i++ {
		key, err := keymap.GetKey("BASE", i)
		if err != nil {
			fmt.Printf("  [%d]: Error - %v\n", i, err)
			continue
		}

		display := key.GetDisplayText()
		if key.IsHoldTap() {
			fmt.Printf("  [%d]: %s (tap: %s, hold: %s)\n", i, display, key.T, key.H)
		} else if key.Type != "" {
			fmt.Printf("  [%d]: %s (type: %s)\n", i, display, key.Type)
		} else {
			fmt.Printf("  [%d]: %s\n", i, display)
		}
	}

	// Show combos
	fmt.Printf("\n🤝 Combos:\n")
	for i, combo := range keymap.Combos {
		fmt.Printf("  [%d]: positions %v → %s (layers: %v)\n", i, combo.P, combo.K, combo.L)
	}

	// Test different layers
	fmt.Printf("\n🔍 Testing layer differences at position 0:\n")
	for _, layer := range keymap.GetLayerNames() {
		key, err := keymap.GetKey(layer, 0)
		if err != nil {
			continue
		}
		fmt.Printf("  %s: %s\n", layer, key.GetDisplayText())
	}

	fmt.Printf("\n✅ Keymap parsing successful!\n")
}