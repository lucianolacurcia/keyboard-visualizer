package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Physical layout definitions
type PhysicalLayout struct {
	Url     string                      `json:"url"`
	Layouts map[string]PhysicalKeyboard `json:"layouts"`
}

type PhysicalKeyboard struct {
	Layout []KeyPosition `json:"layout"`
}

type KeyPosition struct {
	X  float32 `json:"x"`  // X coordinate (in key units)
	Y  float32 `json:"y"`  // Y coordinate (in key units)
	W  float32 `json:"w"`  // Width (in key units, default 1.0)
	H  float32 `json:"h"`  // Height (in key units, default 1.0)
	R  float32 `json:"r"`  // Rotation angle (degrees)
	RX float32 `json:"rx"` // Rotation center X
	RY float32 `json:"ry"` // Rotation center Y
}

// Visual state for each key
type KeyState struct {
	IsPressed     bool   // Currently being pressed
	CurrentLayer  string // Which layer is currently active
	PressedTime   int64  // When key was pressed (for animations)
	IsVisible     bool   // Whether key should be shown in current layer
	IsTransparent bool   // Whether key is transparent in current layer
	IsHeld        bool   // Whether key is in "held" state
}

// Complete drawing information for a key
type KeyDrawInfo struct {
	Position     KeyPosition // Physical position and rotation
	LogicalIndex int         // Index in keymap (0-37 for Totem)

	// Visual coordinates (converted to pixels)
	PixelX      float32
	PixelY      float32
	PixelWidth  float32
	PixelHeight float32
	RotationDeg float32
	RotationCenterX float32
	RotationCenterY float32

	// Key content
	KeyEntry *KeyEntry // From parsed keymap
	State    KeyState  // Current visual state

	// Colors and visual properties
	BackgroundColor rl.Color
	TextColor       rl.Color
	BorderColor     rl.Color
	FontSize        int32
}

// Rendering configuration
type RenderConfig struct {
	KeyUnitSize     float32 // Pixels per key unit (1u = 60px typically)
	KeyPadding      float32 // Inner padding of key
	FontSize        int32   // Font size for key labels
	BorderWidth     float32 // Border thickness
	WindowWidth     int32   // Window width
	WindowHeight    int32   // Window height
	OffsetX         float32 // Global X offset
	OffsetY         float32 // Global Y offset

	// Colors
	IdleColor       rl.Color // Key color when idle
	PressedColor    rl.Color // Key color when pressed
	TransparentColor rl.Color // Key color when transparent
	TextColor       rl.Color // Text color
	BorderColor     rl.Color // Border color
}

// Complete keyboard visualizer state
type KeyboardVisualizer struct {
	Keymap         *KeymapData
	PhysicalLayout *PhysicalLayout
	Keys           []KeyDrawInfo
	Config         RenderConfig
	CurrentLayer   string
	LayerStack     []string // For layer management
}

// LoadPhysicalLayout loads the physical layout JSON
func LoadPhysicalLayout(filename string) (*PhysicalLayout, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read layout file: %w", err)
	}

	var layout PhysicalLayout
	if err := json.Unmarshal(data, &layout); err != nil {
		return nil, fmt.Errorf("failed to parse layout JSON: %w", err)
	}

	return &layout, nil
}

// NewKeyboardVisualizer creates a new keyboard visualizer
func NewKeyboardVisualizer(keymapFile, layoutFile string) (*KeyboardVisualizer, error) {
	// Load keymap
	keymap, err := LoadKeymap(keymapFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load keymap: %w", err)
	}

	// Load physical layout
	physicalLayout, err := LoadPhysicalLayout(layoutFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load physical layout: %w", err)
	}

	// Calculate window size based on keyboard layout
	windowWidth, windowHeight := calculateWindowSize(physicalLayout)

	// Create render config with calculated size
	config := RenderConfig{
		KeyUnitSize:      60.0,
		KeyPadding:       4.0,
		FontSize:         14,
		BorderWidth:      1.0,
		WindowWidth:      windowWidth,
		WindowHeight:     windowHeight,
		OffsetX:          30.0, // Reduced for better fit
		OffsetY:          30.0,
		IdleColor:        rl.NewColor(64, 64, 64, 255),     // Dark gray
		PressedColor:     rl.NewColor(79, 195, 247, 255),   // Cyan
		TransparentColor: rl.NewColor(64, 64, 64, 128),     // Semi-transparent gray
		TextColor:        rl.White,
		BorderColor:      rl.NewColor(128, 128, 128, 255),
	}

	visualizer := &KeyboardVisualizer{
		Keymap:         keymap,
		PhysicalLayout: physicalLayout,
		Config:         config,
		CurrentLayer:   "BASE",
		LayerStack:     []string{"BASE"},
	}

	// Initialize key draw info
	if err := visualizer.initializeKeys(); err != nil {
		return nil, fmt.Errorf("failed to initialize keys: %w", err)
	}

	return visualizer, nil
}

// initializeKeys combines keymap data with physical layout
func (kv *KeyboardVisualizer) initializeKeys() error {
	layout, exists := kv.PhysicalLayout.Layouts["LAYOUT"]
	if !exists {
		return fmt.Errorf("LAYOUT not found in physical layout")
	}

	keyCount := kv.Keymap.GetKeyCount()
	if len(layout.Layout) != keyCount {
		return fmt.Errorf("physical layout has %d keys but keymap has %d", len(layout.Layout), keyCount)
	}

	kv.Keys = make([]KeyDrawInfo, keyCount)

	for i := 0; i < keyCount; i++ {
		pos := layout.Layout[i]

		// Get key entry for current layer
		keyEntry, err := kv.Keymap.GetKey(kv.CurrentLayer, i)
		if err != nil {
			return fmt.Errorf("failed to get key %d for layer %s: %w", i, kv.CurrentLayer, err)
		}

		// Convert units to pixels
		pixelX := pos.X * kv.Config.KeyUnitSize + kv.Config.OffsetX
		pixelY := pos.Y * kv.Config.KeyUnitSize + kv.Config.OffsetY
		pixelWidth := pos.W * kv.Config.KeyUnitSize
		pixelHeight := pos.W * kv.Config.KeyUnitSize // Use width for square keys
		if pos.H > 0 {
			pixelHeight = pos.H * kv.Config.KeyUnitSize
		}

		// Convert rotation center
		rotationCenterX := pos.RX * kv.Config.KeyUnitSize + kv.Config.OffsetX
		rotationCenterY := pos.RY * kv.Config.KeyUnitSize + kv.Config.OffsetY

		kv.Keys[i] = KeyDrawInfo{
			Position:        pos,
			LogicalIndex:    i,
			PixelX:          pixelX,
			PixelY:          pixelY,
			PixelWidth:      pixelWidth,
			PixelHeight:     pixelHeight,
			RotationDeg:     pos.R,
			RotationCenterX: rotationCenterX,
			RotationCenterY: rotationCenterY,
			KeyEntry:        keyEntry,
			State: KeyState{
				IsPressed:     false,
				CurrentLayer:  kv.CurrentLayer,
				IsVisible:     true,
				IsTransparent: keyEntry.Type == "trans",
				IsHeld:        keyEntry.Type == "held",
			},
			BackgroundColor: kv.getKeyColor(keyEntry),
			TextColor:       kv.Config.TextColor,
			BorderColor:     kv.Config.BorderColor,
			FontSize:        kv.Config.FontSize,
		}
	}

	return nil
}

// getKeyColor returns the appropriate color for a key based on its type
func (kv *KeyboardVisualizer) getKeyColor(key *KeyEntry) rl.Color {
	switch key.Type {
	case "trans":
		return kv.Config.TransparentColor
	case "held":
		return kv.Config.PressedColor
	default:
		return kv.Config.IdleColor
	}
}

// SwitchLayer changes the current layer and updates all keys
func (kv *KeyboardVisualizer) SwitchLayer(layer string) error {
	// Verify layer exists
	layerKeys, exists := kv.Keymap.Layers[layer]
	if !exists {
		return fmt.Errorf("layer %s does not exist", layer)
	}

	kv.CurrentLayer = layer

	// Update all key entries and states
	for i := 0; i < len(kv.Keys); i++ {
		if i < len(layerKeys) {
			kv.Keys[i].KeyEntry = &layerKeys[i]
			kv.Keys[i].State.CurrentLayer = layer
			kv.Keys[i].State.IsTransparent = layerKeys[i].Type == "trans"
			kv.Keys[i].State.IsHeld = layerKeys[i].Type == "held"
			kv.Keys[i].BackgroundColor = kv.getKeyColor(&layerKeys[i])
		}
	}

	return nil
}

// PressKey marks a key as pressed (for visual feedback)
func (kv *KeyboardVisualizer) PressKey(keyIndex int) {
	if keyIndex >= 0 && keyIndex < len(kv.Keys) {
		kv.Keys[keyIndex].State.IsPressed = true
		kv.Keys[keyIndex].BackgroundColor = kv.Config.PressedColor
	}
}

// ReleaseKey marks a key as released
func (kv *KeyboardVisualizer) ReleaseKey(keyIndex int) {
	if keyIndex >= 0 && keyIndex < len(kv.Keys) {
		kv.Keys[keyIndex].State.IsPressed = false
		kv.Keys[keyIndex].BackgroundColor = kv.getKeyColor(kv.Keys[keyIndex].KeyEntry)
	}
}

// calculateWindowSize computes optimal window size based on keyboard layout
func calculateWindowSize(physicalLayout *PhysicalLayout) (int32, int32) {
	layout, exists := physicalLayout.Layouts["LAYOUT"]
	if !exists {
		return 900, 400 // fallback
	}

	if len(layout.Layout) == 0 {
		return 900, 400 // fallback
	}

	keyUnitSize := float32(60.0)
	padding := float32(60.0) // padding around the keyboard

	// Find bounds of all keys
	var minX, maxX, minY, maxY float32
	minX = layout.Layout[0].X
	maxX = layout.Layout[0].X + layout.Layout[0].W
	minY = layout.Layout[0].Y
	maxY = layout.Layout[0].Y

	for _, key := range layout.Layout {
		// Calculate key bounds
		keyWidth := key.W
		if keyWidth == 0 {
			keyWidth = 1.0 // default width
		}
		keyHeight := key.H
		if keyHeight == 0 {
			keyHeight = 1.0 // default height
		}

		// Update bounds
		if key.X < minX {
			minX = key.X
		}
		if key.X+keyWidth > maxX {
			maxX = key.X + keyWidth
		}
		if key.Y < minY {
			minY = key.Y
		}
		if key.Y+keyHeight > maxY {
			maxY = key.Y + keyHeight
		}
	}

	// Calculate window size
	width := int32((maxX-minX)*keyUnitSize + 2*padding)
	height := int32((maxY-minY)*keyUnitSize + 2*padding)

	// Ensure minimum size
	if width < 400 {
		width = 400
	}
	if height < 200 {
		height = 200
	}

	return width, height
}

// GetWindowSize returns the configured window size
func (kv *KeyboardVisualizer) GetWindowSize() (int32, int32) {
	return kv.Config.WindowWidth, kv.Config.WindowHeight
}