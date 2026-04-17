package hid

import (
	"fmt"
	"log"
	"time"

	"github.com/sstallion/go-hid"
)

// Reader manages HID device connection and reading
type Reader struct {
	eventChan chan<- *HIDEvent
	stopChan  chan bool
}

// NewReader creates a new HID reader
func NewReader(eventChan chan<- *HIDEvent) *Reader {
	return &Reader{
		eventChan: eventChan,
		stopChan:  make(chan bool),
	}
}

// Start begins reading HID events in a goroutine
func (r *Reader) Start() error {
	// Initialize HID subsystem
	if err := hid.Init(); err != nil {
		return fmt.Errorf("hid init error: %w", err)
	}

	go r.readLoop()
	return nil
}

// Stop stops the HID reader
func (r *Reader) Stop() {
	select {
	case r.stopChan <- true:
	default:
	}
}

// readLoop is the main reading loop that runs in a goroutine
func (r *Reader) readLoop() {
	defer hid.Exit()

	for {
		select {
		case <-r.stopChan:
			log.Println("HID reader stopping...")
			return
		default:
			// Try to connect and read
			if err := r.connectAndRead(); err != nil {
				log.Printf("HID connection error: %v", err)
				log.Println("Retrying in 2 seconds...")
				time.Sleep(2 * time.Second)
			}
		}
	}
}

// connectAndRead attempts to connect to HID device and read events
func (r *Reader) connectAndRead() error {
	// Find Raw HID device
	devicePath, err := r.findRawHIDDevice()
	if err != nil {
		return err
	}

	// Open device
	device, err := hid.OpenPath(devicePath)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	defer device.Close()

	log.Printf("Connected to Raw HID device: %s", devicePath)

	// Read loop
	buf := make([]byte, ReportSize)
	for {
		select {
		case <-r.stopChan:
			return nil
		default:
			n, err := device.Read(buf)
			if err != nil {
				return fmt.Errorf("read error: %w", err)
			}

			if n > 0 {
				if event, parseErr := ParseReport(buf[:n]); parseErr == nil {
					select {
					case r.eventChan <- event:
					case <-r.stopChan:
						return nil
					}
				} else {
					log.Printf("Parse error: %v", parseErr)
				}
			}
		}
	}
}

// findRawHIDDevice searches for and returns the path of a Raw HID device
func (r *Reader) findRawHIDDevice() (string, error) {
	var targetPath string

	err := hid.Enumerate(0, 0, func(info *hid.DeviceInfo) error {
		if info.UsagePage == RawHIDUsagePage {
			log.Printf("Found Raw HID device: %s (VID: 0x%04X, PID: 0x%04X)",
				info.ProductStr, info.VendorID, info.ProductID)
			targetPath = info.Path
			return nil // Found it, but continue enumeration to see all devices
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to enumerate devices: %w", err)
	}

	if targetPath == "" {
		return "", fmt.Errorf("no Raw HID device found (usage page 0x%04X)", RawHIDUsagePage)
	}

	return targetPath, nil
}