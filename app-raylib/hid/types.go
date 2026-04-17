package hid

// HIDEventType defines the type of HID event
type HIDEventType uint8

const (
	EventTypeKey   HIDEventType = 0x01
	EventTypeLayer HIDEventType = 0x02
)

// HIDEvent represents a unified event from the keyboard
type HIDEvent struct {
	Type HIDEventType
	Data interface{}
}

// KeyEvent represents a key press/release event
type KeyEvent struct {
	Position uint16
	Pressed  bool
}

// LayerEvent represents a layer activation/deactivation event
type LayerEvent struct {
	Layer  uint8
	Active bool
}

// Raw HID constants
const (
	RawHIDUsagePage = 0xFF60
	ReportSize      = 32
)