# Go Package Documentation: `soreaja-thermal`

`github.com/punkestu/soreaja-thermal` is a modular, high-performance Go library and toolset for controlling ESC/POS thermal receipt printers over Bluetooth (RFCOMM/SPP) and other transports.

---

## Table of Contents

1. [Package soreaja (Root Facade)](#package-soreaja)
   - [Overview & Quick Start](#quick-start)
   - [Constants & Variables](#constants--variables)
   - [Functions](#root-functions)
   - [Types](#root-types)
2. [Package pkg/printer](#package-pkgprinter)
   - [Client Struct & Methods](#client-struct)
   - [Connection & Transport Management](#connection-management)
   - [Formatting & Printing Methods](#formatting--printing-methods)
3. [Package pkg/escpos](#package-pkgescpos)
   - [Command Generators](#escpos-commands)
   - [Floyd-Steinberg Image Rasterizer](#image-rasterization)
   - [QR Code Generator](#qr-code-generation)
4. [Package pkg/transport](#package-pkgtransport)
   - [Interfaces (Transport, Provider)](#transport-interfaces)
   - [macOS Darwin Driver](#darwin-provider)
   - [Mock Testing Transport](#mock-transport)

---

## Package `soreaja`

```go
import "github.com/punkestu/soreaja-thermal"
```

The root `soreaja` package serves as an ergonomic entry point and facade, re-exporting the primary `Client`, device discovery functions, alignment constants, and common errors from the underlying `pkg/` packages.

### Quick Start

```go
package main

import (
	"log"

	soreaja "github.com/punkestu/soreaja-thermal"
)

func main() {
	// 1. Discover paired devices
	devices, err := soreaja.GetPairedDevices()
	if err != nil || len(devices) == 0 {
		log.Fatalf("No paired printer found: %v", err)
	}

	// 2. Connect to the printer (RFCOMM channel 1 by default)
	client, err := soreaja.Connect(devices[0].Address)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// 3. Format and print a receipt
	_ = client.Init()
	_ = client.SetAlign(soreaja.AlignCenter)
	_ = client.SetFontSize(1, 1) // 2x width, 2x height
	_ = client.SetBold(true)
	_ = client.PrintLine("COFFEE SHOP")

	_ = client.SetFontSize(0, 0)
	_ = client.SetBold(false)
	_ = client.SetAlign(soreaja.AlignLeft)
	_ = client.PrintLine("Espresso Single           $3.50")
	_ = client.PrintLine("Oat Milk Latte            $5.00")
	_ = client.PrintLine("--------------------------------")
	_ = client.SetBold(true)
	_ = client.PrintLine("TOTAL                     $8.50")
	_ = client.SetBold(false)

	// 4. Print QR code and cut paper
	_ = client.SetAlign(soreaja.AlignCenter)
	_ = client.PrintQR("https://receipts.example.com/tx/98712", 6)
	_ = client.CutPaper(3)
}
```

### Constants & Variables

```go
const (
	AlignLeft   Alignment = 0 // Left-justified text
	AlignCenter Alignment = 1 // Centered text
	AlignRight  Alignment = 2 // Right-justified text
)

var (
	ErrNotConnected     = errors.New("printer is not connected")
	ErrAlreadyConnected = errors.New("printer is already connected")
	ErrConnectionFailed = errors.New("failed to connect to RFCOMM channel")
)
```

### Root Functions

#### `Connect`
```go
func Connect(address string, channelID ...uint8) (*Client, error)
```
Connects directly to a thermal printer at the given Bluetooth MAC address or hardware identifier. An optional `channelID` can be provided (default is `1`).

#### `NewClient`
```go
func NewClient(provider ...transport.Provider) *Client
```
Instantiates a new printer `Client`. If no custom provider is passed, the host operating system's default Bluetooth provider (e.g., macOS `DarwinProvider`) is used.

#### `ScanDevices`
```go
func ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error)
```
Executes a 5-second Bluetooth inquiry for nearby devices. Invokes optional `onFound` callbacks as each device is discovered.

#### `GetPairedDevices`
```go
func GetPairedDevices() ([]DeviceInfo, error)
```
Retrieves all Bluetooth devices currently paired with the host system.

### Root Types

```go
type Client = printer.Client
type DeviceInfo = transport.DeviceInfo
type Alignment = escpos.Alignment
```

---

## Package `pkg/printer`

```go
import "github.com/punkestu/soreaja-thermal/pkg/printer"
```

`printer` provides the high-level `Client` struct that manages connections, text layout state, font manipulation, image slicing, and receipt diagnostic testing.

### Client Struct

```go
type Client struct {
    // Unexported connection and synchronization fields
}
```

### Connection Management

#### `func (c *Client) Connect(address string, channelID uint8) error`
Opens an RFCOMM channel to the target Bluetooth address on the given channel ID.

#### `func (c *Client) Disconnect() error`
Closes the underlying transport channel and releases hardware resources.

#### `func (c *Client) IsConnected() bool`
Reports whether the active transport link is connected and healthy.

#### `func (c *Client) SendRaw(data []byte) error`
Transmits raw byte slices directly over the transport link to the printer.

### Formatting & Printing Methods

| Method | Signature | Description |
|---|---|---|
| `Init` | `Init() error` | Resets printer hardware to default power-on state (`ESC @`). |
| `LineFeed` | `LineFeed() error` | Transmits a single line feed byte (`0x0A`). |
| `Feed` | `Feed(lines uint8) error` | Feeds paper forward by `lines` count (`ESC d n`). |
| `CutPaper` | `CutPaper(feedLines uint8) error` | Feeds `feedLines` and cuts paper (`GS V B n`). |
| `SetAlign` | `SetAlign(align Alignment) error` | Sets horizontal alignment (`AlignLeft`, `AlignCenter`, `AlignRight`). |
| `SetBold` | `SetBold(enable bool) error` | Toggles bold/emphasized text (`ESC E n`). |
| `SetUnderline` | `SetUnderline(enable bool) error` | Toggles single-underline text (`ESC - n`). |
| `SetFontSize` | `SetFontSize(width, height uint8) error` | Sets character font scale multiplier (0-7 width, 0-7 height). |
| `PrintText` | `PrintText(text string) error` | Sends text directly without appending a newline. |
| `PrintLine` | `PrintLine(text string) error` | Sends text followed immediately by a newline. |
| `PrintQR` | `PrintQR(text string, size uint8) error` | Generates and prints an ESC/POS Model 2 QR code (`size` 1-16). |
| `PrintImage` | `PrintImage(filePath string, maxWidth int, dither bool) error` | Decodes, dithers, and prints a graphic image file. |
| `PrintTestReceipt` | `PrintTestReceipt(widthDots int) error` | Prints a comprehensive multi-feature diagnostic receipt. |

---

## Package `pkg/escpos`

```go
import "github.com/punkestu/soreaja-thermal/pkg/escpos"
```

`escpos` implements low-level, pure-Go ESC/POS byte sequence generators and image rasterization algorithms without external hardware dependencies.

### ESC/POS Commands

- **`Init() []byte`**: Returns `[]byte{0x1B, 0x40}` (`ESC @`).
- **`LineFeed() []byte`**: Returns `[]byte{0x0A}` (`LF`).
- **`Feed(lines uint8) []byte`**: Returns `[]byte{0x1B, 0x64, lines}` (`ESC d n`).
- **`Cut(feedLines uint8) []byte`**: Feeds `feedLines` and performs paper cut (`GS V B 0`).
- **`Align(alignment Alignment) []byte`**: Returns `[]byte{0x1B, 0x61, byte(alignment)}` (`ESC a n`).
- **`Bold(enable bool) []byte`**: Returns `[]byte{0x1B, 0x45, n}` (`ESC E n`).
- **`Underline(enable bool) []byte`**: Returns `[]byte{0x1B, 0x2D, n}` (`ESC - n`).
- **`Invert(enable bool) []byte`**: Returns `[]byte{0x1D, 0x42, n}` (`GS B n`).
- **`FontSize(width, height uint8) []byte`**: Returns `[]byte{0x1D, 0x21, n}` (`GS ! n`).
- **`CashDrawer() []byte`**: Generates a standard 25ms pulse on pin 2 (`ESC p 0 25 250`).

### Image Rasterization

#### `func RasterizeFile(path string, maxWidth int, dither bool) ([][]byte, error)`
Reads a PNG or JPEG file from disk and converts it into ESC/POS raster chunks.

#### `func RasterizeReader(r io.Reader, maxWidth int, dither bool) ([][]byte, error)`
Reads an image stream from an `io.Reader` and converts it into raster chunks.

#### `func RasterizeImage(src image.Image, maxWidth int, dither bool, chunkHeight int) [][]byte`
Core rasterizer converting any Go `image.Image` into sliced chunks of `GS v 0 0` commands.
- **`maxWidth`**: Standard receipt printer dot widths:
  - `384` dots for **58mm** printers.
  - `576` dots for **80mm** printers.
- **`dither`**: When `true`, applies **Floyd-Steinberg error diffusion** across rows and columns to generate smooth gradients for photographs and logos.
- **`chunkHeight`**: Maximum vertical dot rows per chunk (default `128`) to prevent thermal print head buffer overflows.

### QR Code Generation

#### `func QRCode(text string, size uint8) []byte`
Constructs native ESC/POS Model 2 QR code commands (`GS ( k ...`):
1. Select Model: Model 2 (`0x32`).
2. Module Size: `1` to `16` dots per module.
3. Error Correction: Level `M` (~15% recovery capacity).
4. Store Data & Print Command sequence.

---

## Package `pkg/transport`

```go
import "github.com/punkestu/soreaja-thermal/pkg/transport"
```

`transport` isolates I/O communication from receipt formatting logic, enabling cross-platform adaptability and unit testing.

### Transport Interfaces

```go
// Transport represents an active bidirectional data link.
type Transport interface {
	io.Writer
	io.Closer
	IsConnected() bool
}

// Provider represents a platform Bluetooth discovery and connection factory.
type Provider interface {
	ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error)
	GetPairedDevices() ([]DeviceInfo, error)
	Connect(address string, channelID uint8) (Transport, error)
}
```

### Darwin Provider (macOS)

Built with CGO when compiling for macOS (`//go:build darwin && cgo`):
- Bridges Go with `bridge/mac/libprinterbridge.dylib` via C-ABI.
- Interfaces with macOS `IOBluetooth` framework for hardware inquiry and RFCOMM channel streaming.

### Mock Transport (Testing)

```go
mockProv := transport.NewMockProvider()
client := soreaja.NewClient(mockProv)
_ = client.Connect("00:11:22:33:44:55", 1)

// Written bytes can be inspected directly:
printedBytes := mockProv.CurrentTrans.Buffer.Bytes()
```

---

## CLI & Tooling

The library includes CLI and UI binaries:
- `bin/printer-cli paired` : List paired Bluetooth printers.
- `bin/printer-cli scan` : Discover nearby devices.
- `bin/printer-cli test <mac> [width]` : Run full hardware self-test print.
- `bin/printer-cli text <mac> "<text>"` : Print text snippet.
- `bin/printer-cli image <mac> <file.png> [width]` : Print graphic with dithering.
- `bin/printer-cli raw <mac> "<hex>"` : Send raw ESC/POS bytes.
