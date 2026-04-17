package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sstallion/go-hid"
)

// HID communication constants
const (
	VIA_USAGE_PAGE = 0xff60
	REPORT_SIZE    = 32
)

// KeyPeek-compatible protocol types
const (
	REPORT_TYPE_LAYER_STATE = 0xff // Complete layer state (stateless)
	REPORT_TYPE_KEY_EVENT   = 0xF1 // Individual key events (eventful)
)

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
	device      *hid.Device
	eventChan   chan HIDEvent
	stopChan    chan bool
	isConnected bool
}

// NewHIDReader creates a new HID reader
func NewHIDReader() *HIDReader {
	return &HIDReader{
		eventChan: make(chan HIDEvent, 100),
		stopChan:  make(chan bool),
	}
}

// FindKeyboardDevice finds ZMK Raw HID device
func (hr *HIDReader) FindKeyboardDevice() error {
	if err := hid.Init(); err != nil {
		return fmt.Errorf("failed to initialize HID: %w", err)
	}

	// Enumerate HID devices with callback
	var foundDevice *hid.DeviceInfo

	err := hid.Enumerate(0, 0, func(info *hid.DeviceInfo) error {
		if info.UsagePage == VIA_USAGE_PAGE {
			log.Printf("Found ZMK device: %04X:%04X (Usage: %04X)",
				info.VendorID, info.ProductID, info.UsagePage)
			foundDevice = info
			return fmt.Errorf("device found") // Stop enumeration
		}
		return nil
	})

	if foundDevice == nil {
		return fmt.Errorf("no ZMK Raw HID device found")
	}

	// Try to open device
	device, err := hid.Open(foundDevice.VendorID, foundDevice.ProductID, "")
	if err != nil {
		return fmt.Errorf("failed to open device %04X:%04X: %w",
			foundDevice.VendorID, foundDevice.ProductID, err)
	}

	hr.device = device
	hr.isConnected = true
	return nil
}

// StartReading begins reading HID reports in a goroutine
func (hr *HIDReader) StartReading() error {
	if !hr.isConnected {
		return fmt.Errorf("device not connected")
	}

	go hr.readLoop()
	log.Println("HID reader started")
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
		hr.device.Close()
		hr.isConnected = false
		log.Println("HID reader stopped")
	}
}

// IsConnected returns true if HID device is connected
func (hr *HIDReader) IsConnected() bool {
	return hr.isConnected
}

// readLoop continuously reads HID reports
func (hr *HIDReader) readLoop() {
	buffer := make([]byte, REPORT_SIZE)

	for {
		select {
		case <-hr.stopChan:
			return
		default:
			// Read HID report (blocking read)
			n, err := hr.device.Read(buffer)
			if err != nil {
				log.Printf("HID read error: %v", err)
				time.Sleep(10 * time.Millisecond)
				continue
			}

			if n == 0 {
				// No data, continue reading
				continue
			}

			// Parse and send event
			if event := hr.parseReport(buffer[:n]); event != nil {
				select {
				case hr.eventChan <- *event:
					// Event sent successfully
				default:
					// Channel full, drop event
					log.Println("Warning: event channel full, dropping event")
				}
			}
		}
	}
}

// parseReport parses HID report into events
func (hr *HIDReader) parseReport(data []byte) *HIDEvent {
	if len(data) < REPORT_SIZE {
		return nil
	}

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
	if layerState == 0 || layerState == 1 {
		return "BASE"
	}

	// Check individual layer bits (assuming layers 0=BASE, 1=WM, 2=WMS)
	if layerState&(1<<2) != 0 {
		return "WMS"
	}
	if layerState&(1<<1) != 0 {
		return "WM"
	}

	return "BASE" // fallback
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