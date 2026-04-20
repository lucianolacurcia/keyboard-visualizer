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
type KeyEntry struct {
	// Simple key (when just a string like "Q", "W", etc.)
	Simple string

	// Complex key properties
	T    string `yaml:"t"`    // tap behavior
	H    string `yaml:"h"`    // hold behavior
	Type string `yaml:"type"` // trans, held, ghost, etc.
}

// Combo represents key combinations
type Combo struct {
	P []int    `yaml:"p"` // positions (key indices)
	K KeyEntry `yaml:"k"` // key output (can be simple string or complex object)
	L []string `yaml:"l"` // active layers
}

// UnmarshalYAML handles both string and object key entries
func (k *KeyEntry) UnmarshalYAML(value *yaml.Node) error {
	// Try to unmarshal as string first
	var simple string
	if err := value.Decode(&simple); err == nil {
		k.Simple = simple
		return nil
	}

	// If that fails, try as complex object
	type keyEntryAlias KeyEntry
	var complex keyEntryAlias
	if err := value.Decode(&complex); err != nil {
		return err
	}

	k.T = complex.T
	k.H = complex.H
	k.Type = complex.Type
	return nil
}

// GetDisplayText returns the text to display for this key
func (k *KeyEntry) GetDisplayText() string {
	if k.Simple != "" {
		return k.Simple
	}
	if k.T != "" {
		return k.T // Show tap behavior as main text
	}
	if k.Type == "trans" {
		return "▽" // Transparent key indicator
	}
	if k.Type == "held" {
		return "●" // Held key indicator
	}
	return ""
}

// IsHoldTap returns true if this key has hold-tap behavior
func (k *KeyEntry) IsHoldTap() bool {
	return k.T != "" && k.H != ""
}

// GetHoldBehavior returns the hold behavior if this is a hold-tap key
func (k *KeyEntry) GetHoldBehavior() string {
	return k.H
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