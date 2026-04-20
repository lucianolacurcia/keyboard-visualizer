package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sstallion/go-hid"
)

// HID communication constants
const (
	VIA_USAGE_PAGE      = 0xff60 // Our custom protocol
	KEYBOARD_USAGE_PAGE = 0x0001 // Standard keyboard protocol
	REPORT_SIZE         = 32
)

// KeyPeek-compatible protocol types
const (
	REPORT_TYPE_LAYER_STATE = 0xff // Complete layer state (stateless)
	REPORT_TYPE_KEY_EVENT   = 0xF1 // Individual key events (eventful)
)

// USB HID Keyboard keycode mapping
var hidKeycodes = map[uint8]string{
	0x00: "", // No key
	0x04: "A", 0x05: "B", 0x06: "C", 0x07: "D", 0x08: "E", 0x09: "F", 0x0A: "G",
	0x0B: "H", 0x0C: "I", 0x0D: "J", 0x0E: "K", 0x0F: "L", 0x10: "M", 0x11: "N",
	0x12: "O", 0x13: "P", 0x14: "Q", 0x15: "R", 0x16: "S", 0x17: "T", 0x18: "U",
	0x19: "V", 0x1A: "W", 0x1B: "X", 0x1C: "Y", 0x1D: "Z",
	0x1E: "1", 0x1F: "2", 0x20: "3", 0x21: "4", 0x22: "5",
	0x23: "6", 0x24: "7", 0x25: "8", 0x26: "9", 0x27: "0",
	0x28: "ENTER", 0x29: "ESC", 0x2A: "BSPC", 0x2B: "TAB", 0x2C: "SPACE",
	0x2D: "-", 0x2E: "=", 0x2F: "[", 0x30: "]", 0x31: "\\", 0x33: ";", 0x34: "'",
	0x35: "`", 0x36: ",", 0x37: ".", 0x38: "/",
	0x39: "CAPS", 0x3A: "F1", 0x3B: "F2", 0x3C: "F3", 0x3D: "F4", 0x3E: "F5",
	0x3F: "F6", 0x40: "F7", 0x41: "F8", 0x42: "F9", 0x43: "F10", 0x44: "F11", 0x45: "F12",
	0x4C: "DEL", 0x4F: "RIGHT", 0x50: "LEFT", 0x51: "DOWN", 0x52: "UP",
}

// USB HID Modifier keys (byte 0)
var hidModifiers = map[uint8]string{
	0x01: "LCTRL", 0x02: "LSHIFT", 0x04: "LALT", 0x08: "LGUI",
	0x10: "RCTRL", 0x20: "RSHIFT", 0x40: "RALT", 0x80: "RGUI",
}

// HID Events
type HIDKeyEvent struct {
	Row     uint8
	Col     uint8
	Pressed bool
}

type HIDLayerState struct {
	DefaultLayerState uint32
	LayerState        uint32
}

type HIDEvent struct {
	Type      uint8
	KeyEvent  *HIDKeyEvent
	LayerState *HIDLayerState
}

// HID Reader manages connection and parsing
type HIDReader struct {
	customDevice   *hid.Device // Our custom protocol (0xFF60)
	keyboardDevice *hid.Device // Standard keyboard (0x0001)
	eventChan      chan HIDEvent
	stopChan       chan bool
	isConnected    bool
}

// NewHIDReader creates a new HID reader
func NewHIDReader() *HIDReader {
	return &HIDReader{
		eventChan: make(chan HIDEvent, 100),
		stopChan:  make(chan bool),
	}
}

// FindKeyboardDevice finds both ZMK devices (custom + keyboard)
func (hr *HIDReader) FindKeyboardDevice() error {
	if err := hid.Init(); err != nil {
		return fmt.Errorf("failed to initialize HID: %w", err)
	}

	// Find both devices with same VID/PID but different usage pages
	var customDevice *hid.DeviceInfo
	var keyboardDevice *hid.DeviceInfo

	err := hid.Enumerate(0, 0, func(info *hid.DeviceInfo) error {
		if info.UsagePage == VIA_USAGE_PAGE {
			log.Printf("Found ZMK Custom device: %04X:%04X (Usage: %04X)",
				info.VendorID, info.ProductID, info.UsagePage)
			customDevice = info
		} else if info.VendorID == 0x1D50 && info.ProductID == 0x615E &&
		          info.UsagePage == 0x0001 && info.Usage == 0x0006 {
			log.Printf("Found ZMK Keyboard device: %04X:%04X (Usage: %04X/%04X) Path: %s",
				info.VendorID, info.ProductID, info.UsagePage, info.Usage, info.Path)
			keyboardDevice = info // Take ONLY the real keyboard endpoint
		}
		return nil
	})

	if customDevice == nil {
		return fmt.Errorf("no ZMK Custom HID device found")
	}

	// Open custom device (required)
	custom, err := hid.OpenPath(customDevice.Path)
	if err != nil {
		return fmt.Errorf("failed to open custom device %04X:%04X at %s: %w",
			customDevice.VendorID, customDevice.ProductID, customDevice.Path, err)
	}
	hr.customDevice = custom

	// Open keyboard device (optional)
	if keyboardDevice != nil {
		keyboard, err := hid.OpenPath(keyboardDevice.Path)
		if err != nil {
			log.Printf("Warning: failed to open keyboard device: %v", err)
		} else {
			hr.keyboardDevice = keyboard
			log.Printf("Opened both custom and keyboard devices")
		}
	} else {
		log.Printf("Keyboard device not found - will only monitor custom channel")
	}

	hr.isConnected = true
	return nil
}

// StartReading begins reading HID reports in a goroutine
func (hr *HIDReader) StartReading() error {
	if !hr.isConnected || hr.customDevice == nil {
		return fmt.Errorf("device not connected")
	}

	go hr.readLoop()
	log.Println("HID reader started - monitoring both custom and keyboard channels")
	return nil
}

// GetEventChannel returns the event channel
func (hr *HIDReader) GetEventChannel() <-chan HIDEvent {
	return hr.eventChan
}

// Stop stops the HID reader
func (hr *HIDReader) Stop() {
	if hr.isConnected {
		hr.stopChan <- true
		if hr.customDevice != nil {
			hr.customDevice.Close()
		}
		if hr.keyboardDevice != nil {
			hr.keyboardDevice.Close()
		}
		hr.isConnected = false
		log.Println("HID reader stopped")
	}
}

// IsConnected returns true if HID device is connected
func (hr *HIDReader) IsConnected() bool {
	return hr.isConnected
}

// readLoop continuously reads HID reports from both devices
func (hr *HIDReader) readLoop() {
	// Start reader for custom device
	go hr.readCustomLoop()

	// Start reader for keyboard device (if available)
	if hr.keyboardDevice != nil {
		go hr.readKeyboardLoop()
	}

	// Wait for stop signal
	<-hr.stopChan
}

// readCustomLoop reads from custom HID device (our protocol)
func (hr *HIDReader) readCustomLoop() {
	buffer := make([]byte, REPORT_SIZE)

	for {
		// Read HID report (blocking read)
		n, err := hr.customDevice.Read(buffer)
		if err != nil {
			log.Printf("Custom HID read error: %v", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if n == 0 {
			continue
		}

		// Parse and send event
		if event := hr.parseReport(buffer[:n]); event != nil {
			select {
			case hr.eventChan <- *event:
				// Event sent successfully
			default:
				// Channel full, drop event
				log.Println("Warning: event channel full, dropping custom event")
			}
		}
	}
}

// readKeyboardLoop reads from standard keyboard device
func (hr *HIDReader) readKeyboardLoop() {
	buffer := make([]byte, 8) // Standard keyboard reports are 8 bytes

	for {
		// Read HID report (blocking read)
		n, err := hr.keyboardDevice.Read(buffer)
		if err != nil {
			log.Printf("Keyboard HID read error: %v", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if n == 0 {
			continue
		}

		// Parse and log keyboard data
		parsed := parseKeyboardReport(buffer[:n])
		log.Printf("KEYBOARD: %02x -> %s", buffer[:n], parsed)
	}
}

// parseReport parses HID report into events
func (hr *HIDReader) parseReport(data []byte) *HIDEvent {
	if len(data) < REPORT_SIZE {
		return nil
	}

	// Log raw custom HID data for debugging
	log.Printf("CUSTOM:   %02x", data)

	reportType := data[0]

	switch reportType {
	case REPORT_TYPE_LAYER_STATE:
		return hr.parseLayerState(data)
	case REPORT_TYPE_KEY_EVENT:
		return hr.parseKeyEvent(data)
	default:
		log.Printf("Unknown report type: 0x%02X", reportType)
		return nil
	}
}

// parseLayerState parses complete layer state (stateless)
func (hr *HIDReader) parseLayerState(data []byte) *HIDEvent {
	if len(data) < 10 {
		return nil
	}

	size := data[1] // Should be 4 (uint32_t size)
	if size != 4 {
		log.Printf("Invalid layer state size: %d", size)
		return nil
	}

	// Parse default layer state (little-endian)
	defaultLayerState := uint32(data[2]) |
		uint32(data[3])<<8 |
		uint32(data[4])<<16 |
		uint32(data[5])<<24

	// Parse current layer state (little-endian)
	layerState := uint32(data[6]) |
		uint32(data[7])<<8 |
		uint32(data[8])<<16 |
		uint32(data[9])<<24

	return &HIDEvent{
		Type: REPORT_TYPE_LAYER_STATE,
		LayerState: &HIDLayerState{
			DefaultLayerState: defaultLayerState,
			LayerState:        layerState,
		},
	}
}

// parseKeyEvent parses individual key event (eventful)
func (hr *HIDReader) parseKeyEvent(data []byte) *HIDEvent {
	if len(data) < 4 {
		return nil
	}

	row := data[1]
	col := data[2]
	pressed := data[3] != 0

	return &HIDEvent{
		Type: REPORT_TYPE_KEY_EVENT,
		KeyEvent: &HIDKeyEvent{
			Row:     row,
			Col:     col,
			Pressed: pressed,
		},
	}
}

// GetActiveLayerName converts layer state to layer name
func GetActiveLayerName(layerState uint32) string {
	// Find highest active layer (ZMK layer precedence)
	// Layer definitions: BASE=0, NUMS=1, SYMS=2, WM=3, WMS=4, CONFIG=5

	// Check from highest to lowest priority
	if layerState&(1<<5) != 0 {
		return "CONFIG"
	}
	if layerState&(1<<4) != 0 {
		return "WMS"
	}
	if layerState&(1<<3) != 0 {
		return "WM"
	}
	if layerState&(1<<2) != 0 {
		return "SYMS"
	}
	if layerState&(1<<1) != 0 {
		return "NUMS"
	}

	return "BASE" // default/fallback
}

// parseKeyboardReport decodes USB HID keyboard report into human-readable format
func parseKeyboardReport(data []byte) string {
	if len(data) < 8 {
		return "invalid"
	}

	// Debug: show raw bytes
	debug := fmt.Sprintf("[%02x %02x %02x %02x %02x %02x %02x %02x]",
		data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7])

	// Parse modifier keys (byte 0) - ZMK seems to use byte 0 as status flag, not modifiers
	// Skip modifier parsing for now - ZMK format appears non-standard
	_ = data[0] // Ignore byte 0

	// Parse key codes (bytes 3-8) - ZMK uses offset +1
	var keys []string
	for i := 3; i < 9 && i < len(data); i++ {
		if data[i] != 0 {
			if keyName, exists := hidKeycodes[data[i]]; exists && keyName != "" {
				keys = append(keys, keyName)
			} else {
				keys = append(keys, fmt.Sprintf("0x%02X", data[i]))
			}
		}
	}

	// Build result - only show keycodes (no fake modifiers)
	result := "none"
	if len(keys) > 0 {
		result = strings.Join(keys, "+")
	}

	return fmt.Sprintf("%s %s", debug, result)
}

// RowColToPosition converts row/col to keyboard position index
func RowColToPosition(row, col uint8) int {
	// Inverse of position_to_row_col from ZMK module
	if col < 5 {
		// Left half
		return int(row*5 + col)
	} else {
		// Right half
		return int(row*5 + (col-5) + 19)
	}
}