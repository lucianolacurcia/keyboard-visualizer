package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// KeymapData represents the complete keymap YAML structure
type KeymapData struct {
	Layout Layout                `yaml:"layout"`
	Layers map[string][]KeyEntry `yaml:"layers"`
	Combos []Combo               `yaml:"combos"`
}

// Layout defines the physical keyboard layout
type Layout struct {
	ZmkKeyboard string `yaml:"zmk_keyboard"`
}

// KeyEntry can be either a simple string or a complex key object
// Supports full keymap-drawer specification with all 8 positions and aliases
type KeyEntry struct {
	// Simple key (when just a string like "Q", "W", etc.)
	Simple string

	// All 8 keymap-drawer positions with aliases support
	Center      string `yaml:"center,omitempty"`  // Primary position (aliases: t, tap)
	Hold        string `yaml:"hold,omitempty"`    // Hold behavior (aliases: h, bottom)
	Shifted     string `yaml:"shifted,omitempty"` // Shifted position (alias: s, top)
	Left        string `yaml:"left,omitempty"`    // Left position
	Right       string `yaml:"right,omitempty"`   // Right position
	TopLeft     string `yaml:"tl,omitempty"`      // Top-left position
	TopRight    string `yaml:"tr,omitempty"`      // Top-right position
	BottomLeft  string `yaml:"bl,omitempty"`      // Bottom-left position
	BottomRight string `yaml:"br,omitempty"`      // Bottom-right position

	// Aliases (handled in UnmarshalYAML)
	T      string `yaml:"t,omitempty"`      // Alias for center/tap
	Tap    string `yaml:"tap,omitempty"`    // Alias for center
	H      string `yaml:"h,omitempty"`      // Alias for hold
	Bottom string `yaml:"bottom,omitempty"` // Alias for hold
	S      string `yaml:"s,omitempty"`      // Alias for shifted
	Top    string `yaml:"top,omitempty"`    // Alias for shifted

	// Special keymap-drawer features
	Type  string `yaml:"type,omitempty"`  // trans, held, ghost, etc.
	Glyph string `yaml:"glyph,omitempty"` // SVG glyph like $$mdi:icon$$
}

// Combo represents key combinations
type Combo struct {
	P []int    `yaml:"p"` // positions (key indices)
	K KeyEntry `yaml:"k"` // key output (can be simple string or complex object)
	L []string `yaml:"l"` // active layers
}

// UnmarshalYAML handles both string and object key entries
// Implements full keymap-drawer specification with strict validation
func (k *KeyEntry) UnmarshalYAML(value *yaml.Node) error {
	// Try to unmarshal as simple string first
	var simple string
	if err := value.Decode(&simple); err == nil {
		k.Simple = simple
		return nil
	}

	// If that fails, try as complex object with alias resolution
	type keyEntryAlias KeyEntry
	var complex keyEntryAlias
	if err := value.Decode(&complex); err != nil {
		return fmt.Errorf("FATAL: Invalid KeyEntry format: %v", err)
	}

	// Copy all direct fields
	*k = KeyEntry(complex)

	// STRICT alias resolution with fail-fast validation
	aliasCount := 0

	// Handle center/tap aliases - ONLY one should be specified
	if complex.T != "" {
		if complex.Center != "" || complex.Tap != "" {
			return fmt.Errorf("FATAL: Conflicting aliases for center position: found 't' with 'center' or 'tap'")
		}
		k.Center = complex.T
		aliasCount++
	}
	if complex.Tap != "" {
		if complex.Center != "" {
			return fmt.Errorf("FATAL: Conflicting aliases for center position: found 'tap' with 'center'")
		}
		k.Center = complex.Tap
		aliasCount++
	}

	// Handle hold/bottom aliases - ONLY one should be specified
	if complex.H != "" {
		if complex.Hold != "" || complex.Bottom != "" {
			return fmt.Errorf("FATAL: Conflicting aliases for hold position: found 'h' with 'hold' or 'bottom'")
		}
		k.Hold = complex.H
	}
	if complex.Bottom != "" {
		if complex.Hold != "" {
			return fmt.Errorf("FATAL: Conflicting aliases for hold position: found 'bottom' with 'hold'")
		}
		k.Hold = complex.Bottom
	}

	// Handle shifted/top aliases - ONLY one should be specified
	if complex.S != "" {
		if complex.Shifted != "" || complex.Top != "" {
			return fmt.Errorf("FATAL: Conflicting aliases for shifted position: found 's' with 'shifted' or 'top'")
		}
		k.Shifted = complex.S
	}
	if complex.Top != "" {
		if complex.Shifted != "" {
			return fmt.Errorf("FATAL: Conflicting aliases for shifted position: found 'top' with 'shifted'")
		}
		k.Shifted = complex.Top
	}

	// Clear alias fields to avoid confusion
	k.T = ""
	k.Tap = ""
	k.H = ""
	k.Bottom = ""
	k.S = ""
	k.Top = ""

	return nil
}

// GetDisplayText returns the text to display for this key
// Follows keymap-drawer display priority and special symbols
func (k *KeyEntry) GetDisplayText() string {
	// Simple string takes highest priority
	if k.Simple != "" {
		return k.Simple
	}

	// SVG glyph handling
	if k.Glyph != "" {
		// For now, show glyph name (TODO: SVG rendering support)
		return k.Glyph
	}

	// Special type symbols (keymap-drawer standard)
	switch k.Type {
	case "trans":
		return "▽" // Transparent key indicator
	case "held":
		return "●" // Held key indicator
	case "ghost":
		return "👻" // Ghost key indicator
	}

	// Show center/tap behavior as main text
	if k.Center != "" {
		return k.Center
	}

	// Fallback: no display text
	return ""
}

// IsHoldTap returns true if this key has hold-tap behavior
func (k *KeyEntry) IsHoldTap() bool {
	return k.Center != "" && k.Hold != ""
}

// GetHoldBehavior returns the hold behavior if this is a hold-tap key
func (k *KeyEntry) GetHoldBehavior() string {
	return k.Hold
}

// GetAllPositions returns all defined positions for this key (for debugging/validation)
func (k *KeyEntry) GetAllPositions() map[string]string {
	positions := make(map[string]string)

	if k.Simple != "" {
		positions["simple"] = k.Simple
	}
	if k.Center != "" {
		positions["center"] = k.Center
	}
	if k.Hold != "" {
		positions["hold"] = k.Hold
	}
	if k.Shifted != "" {
		positions["shifted"] = k.Shifted
	}
	if k.Left != "" {
		positions["left"] = k.Left
	}
	if k.Right != "" {
		positions["right"] = k.Right
	}
	if k.TopLeft != "" {
		positions["tl"] = k.TopLeft
	}
	if k.TopRight != "" {
		positions["tr"] = k.TopRight
	}
	if k.BottomLeft != "" {
		positions["bl"] = k.BottomLeft
	}
	if k.BottomRight != "" {
		positions["br"] = k.BottomRight
	}
	if k.Glyph != "" {
		positions["glyph"] = k.Glyph
	}
	if k.Type != "" {
		positions["type"] = k.Type
	}

	return positions
}

// LoadKeymap loads and parses the keymap YAML file
func LoadKeymap(filename string) (*KeymapData, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read keymap file: %w", err)
	}

	var keymap KeymapData
	if err := yaml.Unmarshal(data, &keymap); err != nil {
		return nil, fmt.Errorf("failed to parse keymap YAML: %w", err)
	}

	return &keymap, nil
}

// GetKey returns the key at the given position for the specified layer
func (k *KeymapData) GetKey(layer string, position int) (*KeyEntry, error) {
	layerKeys, exists := k.Layers[layer]
	if !exists {
		return nil, fmt.Errorf("layer %s not found", layer)
	}

	if position < 0 || position >= len(layerKeys) {
		return nil, fmt.Errorf("position %d out of range for layer %s", position, layer)
	}

	return &layerKeys[position], nil
}

// GetLayerNames returns all available layer names
func (k *KeymapData) GetLayerNames() []string {
	var names []string
	for name := range k.Layers {
		names = append(names, name)
	}
	return names
}

// GetKeyCount returns the total number of keys (assumes all layers have same count)
func (k *KeymapData) GetKeyCount() int {
	for _, layer := range k.Layers {
		return len(layer)
	}
	return 0
}
