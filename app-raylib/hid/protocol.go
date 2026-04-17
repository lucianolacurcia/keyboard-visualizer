package hid

import "fmt"

// ParseReport parses a 32-byte HID report into an HIDEvent
func ParseReport(data []byte) (*HIDEvent, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty report")
	}

	switch data[0] {
	case uint8(EventTypeKey):
		return parseKeyEvent(data)
	case uint8(EventTypeLayer):
		return parseLayerEvent(data)
	default:
		return nil, fmt.Errorf("unknown event type: 0x%02X", data[0])
	}
}

// parseKeyEvent parses a key press/release event
// Protocol: data[0]=0x01, data[2-3]=position (uint16 LE), data[4]=state
func parseKeyEvent(data []byte) (*HIDEvent, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("key event data too short")
	}

	position := uint16(data[2]) | uint16(data[3])<<8 // Little-endian
	pressed := data[4] == 1

	return &HIDEvent{
		Type: EventTypeKey,
		Data: KeyEvent{
			Position: position,
			Pressed:  pressed,
		},
	}, nil
}

// parseLayerEvent parses a layer activation/deactivation event
// Protocol: data[0]=0x02, data[2]=layer, data[3]=state
func parseLayerEvent(data []byte) (*HIDEvent, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("layer event data too short")
	}

	layer := data[2]
	active := data[3] == 1

	return &HIDEvent{
		Type: EventTypeLayer,
		Data: LayerEvent{
			Layer:  layer,
			Active: active,
		},
	}, nil
}