# Soreaja Thermal: Go Thermal ESC/POS Printer Library & Apps

A cross-platform **Go library** and desktop **CLI / GUI applications** for ESC/POS thermal receipt printers over Bluetooth (RFCOMM/SPP).

Hardware interaction on macOS is bridged via an isolated, lightweight Swift module (`bridge/mac/`) that exposes a minimal C-ABI, decoupling low-level OS Bluetooth internals from portable Go printing logic. Ready for future OS bridges (Windows Winsock RFCOMM, Linux BlueZ).

---

## Architecture Overview

```
soreaja-thermal/
├── go.mod
├── Makefile
├── soreaja.go                 # Root package facade
├── test.png                   # Sample test receipt image
│
├── bridge/
│   └── mac/                   # macOS-specific Swift Bluetooth bridge (RFCOMM only)
│       ├── Makefile           # Builds libprinterbridge.dylib
│       ├── bridge.h           # C-ABI header
│       └── PrinterBridge.swift# IOBluetooth inquiry & RFCOMM transmission
│
├── pkg/
│   ├── escpos/                # Pure Go ESC/POS protocol engine & image dithering
│   │   ├── commands.go        # Init, line feed, cut, align, bold, font sizes
│   │   ├── qrcode.go          # ESC/POS Model 2 QR code generator
│   │   └── image.go           # Floyd-Steinberg error diffusion & GS v 0 rasterizer
│   │
│   ├── transport/             # Transport abstraction layer
│   │   ├── transport.go       # Transport & Provider interfaces
│   │   ├── darwin_cgo.go      # macOS cgo implementation binding to bridge/mac
│   │   └── mock.go            # In-memory mock transport for unit testing
│   │
│   └── printer/               # High-level thermal printer client API
│       └── printer.go         # Connect, PrintText, PrintImage, PrintQR, PrintTestReceipt
│
└── cmd/
    ├── cli/                   # Go CLI application (replaces CLI.swift)
    │   └── main.go
    └── ui/                    # Go Desktop UI application (Fyne GUI, replaces UI.swift)
        └── main.go
```

---

## Features

- **Decoupled Architecture**: Swift is restricted strictly to macOS Bluetooth RFCOMM hardware transmission. All ESC/POS command formatting, layout logic, and image dithering are implemented in 100% portable Go.
- **Pure Go ESC/POS Protocol Engine**:
  - Printer initialization (`ESC @`).
  - Text formatting: Bold, Underline, Invert, Alignments (Left, Center, Right), Font multipliers (up to 8x).
  - Multi-line feeding and paper cutting (`GS V B 0`).
  - Cash drawer kick pulse (`ESC p 0 25 250`).
  - Model 2 QR code generation (`GS ( k ...`).
- **Floyd-Steinberg Image Rasterizer**:
  - Supports PNG and JPEG graphics natively via Go's standard `image` library.
  - Scales to paper widths: **58mm (384 dots)** or **80mm (576 dots)**.
  - Floyd-Steinberg error diffusion dithering for sharp graphics and photos.
  - Slices tall images into vertical chunks to prevent printer memory buffer overflow.
- **Applications**:
  - **Go CLI (`bin/printer-cli`)**: Terminal command-line tool for scripting and quick tests.
  - **Go Desktop GUI (`bin/printer-ui`)**: Native cross-platform Fyne GUI with device discovery, live text composer, image picker, QR generator, raw hex sender, and activity log.
  - **macOS Bundle (`PrinterBridge.app`)**: Packaged `.app` bundle for desktop launching.

---

## Quick Start

### Build All Artifacts

```bash
make all
```

This compiles:
1. `bridge/mac/libprinterbridge.dylib` (macOS RFCOMM bridge)
2. `bin/printer-cli` (Go CLI application)
3. `bin/printer-ui` (Go Desktop UI application)
4. `PrinterBridge.app` (macOS Application bundle)

### Run Unit Tests

```bash
make test
```

---

## Go Library Usage

Add the module to your Go project:

```bash
go get github.com/punkestu/soreaja-thermal
```

### Basic Example

```go
package main

import (
	"log"

	soreaja "github.com/punkestu/soreaja-thermal"
)

func main() {
	// 1. Scan or list paired devices
	devices, err := soreaja.GetPairedDevices()
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}

	for _, dev := range devices {
		log.Printf("Found: %s (%s)", dev.Name, dev.Address)
	}

	// 2. Connect to printer (MAC address, RFCOMM channel default: 1)
	client, err := soreaja.Connect("86-67-7a-7c-0a-56", 1)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Disconnect()

	// 3. Print formatted receipt
	_ = client.Init()
	_ = client.SetAlign(soreaja.AlignCenter)
	_ = client.SetBold(true)
	_ = client.SetFontSize(1, 1) // 2x font
	_ = client.PrintLine("COFFEE SHOP")

	_ = client.SetBold(false)
	_ = client.SetFontSize(0, 0)
	_ = client.SetAlign(soreaja.AlignLeft)
	_ = client.PrintLine("Espresso                $3.00")
	_ = client.PrintLine("Croissant               $2.50")
	_ = client.PrintLine("--------------------------------")

	// 4. Print QR code
	_ = client.SetAlign(soreaja.AlignCenter)
	_ = client.PrintQR("https://example.com/receipt/123", 6)

	// 5. Feed and cut paper
	_ = client.CutPaper(3)
}
```

### Image Printing with Floyd-Steinberg Dithering

```go
// Print image scaled to 384 dots (58mm width) with Floyd-Steinberg dithering
err := client.PrintImage("logo.png", 384, true)
if err != nil {
    log.Fatalf("Image printing failed: %v", err)
}
_ = client.CutPaper(3)
```

---

## Command-Line Interface (CLI)

Run `bin/printer-cli help` for available commands:

```bash
# 1. List paired Bluetooth devices
./bin/printer-cli paired

# 2. Scan for nearby Bluetooth devices (5s inquiry)
./bin/printer-cli scan

# 3. Print diagnostic test receipt
./bin/printer-cli test 86-67-7a-7c-0a-56 384

# 4. Print custom text
./bin/printer-cli text 86-67-7a-7c-0a-56 "Hello from Go!\nSecond Line\n"

# 5. Print image file with dithering
./bin/printer-cli image 86-67-7a-7c-0a-56 test.png 384

# 6. Send raw hex ESC/POS bytes
./bin/printer-cli raw 86-67-7a-7c-0a-56 "1B40 1B6101 48690A 1D564200"
```

---

## Desktop Graphical Interface (GUI)

Launch the native Go desktop application:

```bash
# Run binary directly:
./bin/printer-ui

# Or open the macOS App bundle:
open PrinterBridge.app
# or
make run-ui
```

### UI Features:
- **Device Picker**: Select from paired devices or run a 5-second inquiry scan for nearby printers.
- **Connection Indicator**: Channel selector, 58mm/80mm paper width options, and live connection status.
- **Text Composer Tab**: Rich receipt formatting (Alignments, Bold, Underline, Font scaling multipliers, feed count).
- **Image Tab**: Image file browser, live image preview, and Floyd-Steinberg dithering toggle.
- **QR Code Tab**: Text/URL input with module size slider (1–16).
- **Raw Hex Tab**: Send raw ESC/POS byte sequences with built-in command presets (Init, Cut, Cash Drawer).
- **Diagnostics Tab**: Full diagnostic test receipt covering styles, alignments, and QR codes.
- **Activity Log**: Real-time timestamped log of Bluetooth connections, data transfers, and errors.

---

## License

MIT
