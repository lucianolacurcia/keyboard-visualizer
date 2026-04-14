package main

import (
	"fmt"
	"os"

	"github.com/sstallion/go-hid"
)

const rawHIDUsagePage = 0xFF60

func main() {
	if err := hid.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "hid init error: %v\n", err)
		os.Exit(1)
	}
	defer hid.Exit()

	// Step 1: Enumerate all HID devices and find Raw HID
	fmt.Println("Enumerating HID devices...")
	fmt.Println()

	var targetPath string

	hid.Enumerate(0, 0, func(info *hid.DeviceInfo) error {
		// Print all devices so we can see what's connected
		fmt.Printf("Device: %s\n", info.ProductStr)
		fmt.Printf("  VID: 0x%04X  PID: 0x%04X\n", info.VendorID, info.ProductID)
		fmt.Printf("  Usage Page: 0x%04X  Usage: 0x%04X\n", info.UsagePage, info.Usage)
		fmt.Printf("  Path: %s\n", info.Path)
		fmt.Println()

		if info.UsagePage == rawHIDUsagePage {
			fmt.Println("  ^^^ RAW HID FOUND ^^^")
			fmt.Println()
			targetPath = info.Path
		}

		return nil
	})

	if targetPath == "" {
		fmt.Println("No Raw HID device found (usage page 0xFF60)")
		fmt.Println("Is your keyboard connected with the visualizer firmware?")
		os.Exit(1)
	}

	// Step 2: Open the device
	device, err := hid.OpenPath(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open device: %v\n", err)
		os.Exit(1)
	}
	defer device.Close()

	fmt.Println("Connected! Reading reports... (Ctrl+C to quit)")
	fmt.Println()

	// Step 3: Read loop
	buf := make([]byte, 32)
	for {
		n, err := device.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			break
		}

		if n > 0 {
			printReport(buf[:n])
		}
	}
}

func printReport(data []byte) {
	if len(data) == 0 {
		return
	}

	switch data[0] {
	case 0x01: // Key state
		pos := uint16(data[2]) | uint16(data[3])<<8
		state := "released"
		if data[4] == 1 {
			state = "pressed"
		}
		fmt.Printf("[KEY]   position=%d  %s\n", pos, state)

	case 0x02: // Layer state
		layer := data[2]
		state := "deactivated"
		if data[3] == 1 {
			state = "activated"
		}
		fmt.Printf("[LAYER] layer=%d  %s\n", layer, state)

	default:
		fmt.Printf("[UNKNOWN] type=0x%02X  raw=%x\n", data[0], data)
	}
}
