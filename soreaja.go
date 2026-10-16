// Package soreaja provides a high-level Go client and ESC/POS thermal printing engine.
//
// soreaja enables cross-platform thermal receipt printing over Bluetooth (RFCOMM/SPP)
// and other transports. It handles ESC/POS command generation, text styling, Floyd-Steinberg
// image dithering, and QR code rendering in pure Go, with hardware I/O bridged to platform-native
// drivers (such as IOBluetooth on macOS).
//
// # Quick Start
//
// Establish a connection and print a receipt:
//
//	package main
//
//	import (
//		"log"
//		"github.com/punkestu/soreaja-thermal"
//	)
//
//	func main() {
//		// Connect to a Bluetooth thermal printer at RFCOMM channel 1
//		client, err := soreaja.Connect("86-67-7A-7C-0A-56")
//		if err != nil {
//			log.Fatalf("Failed to connect: %v", err)
//		}
//		defer client.Disconnect()
//
//		// Initialize and format receipt
//		client.Init()
//		client.SetAlign(soreaja.AlignCenter)
//		client.SetBold(true)
//		client.PrintLine("STORE NAME")
//		client.SetBold(false)
//		client.SetAlign(soreaja.AlignLeft)
//		client.PrintLine("Item 1 ................. $10.00")
//		client.PrintLine("Item 2 .................. $5.00")
//		client.PrintQR("https://example.com/receipt/123", 6)
//		client.CutPaper(3)
//	}
package soreaja

import (
	"github.com/punkestu/soreaja-thermal/pkg/escpos"
	"github.com/punkestu/soreaja-thermal/pkg/printer"
	"github.com/punkestu/soreaja-thermal/pkg/transport"
)

// Client manages communication and command formatting for a thermal printer.
type Client = printer.Client

// DeviceInfo represents a detected or paired Bluetooth device.
type DeviceInfo = transport.DeviceInfo

// Alignment defines horizontal text alignment for ESC/POS commands.
type Alignment = escpos.Alignment

// Text alignment constants for ESC/POS formatting.
const (
	// AlignLeft sets left-justified text alignment.
	AlignLeft Alignment = escpos.AlignLeft
	// AlignCenter sets centered text alignment.
	AlignCenter Alignment = escpos.AlignCenter
	// AlignRight sets right-justified text alignment.
	AlignRight Alignment = escpos.AlignRight
)

// Common transport and connection errors.
var (
	// ErrNotConnected indicates an operation was attempted on a disconnected printer.
	ErrNotConnected = transport.ErrNotConnected
	// ErrAlreadyConnected indicates a connection was attempted while already connected.
	ErrAlreadyConnected = transport.ErrAlreadyConnected
	// ErrConnectionFailed indicates failure to establish an RFCOMM link.
	ErrConnectionFailed = transport.ErrConnectionFailed
)

// Connect establishes a connection to a thermal printer at the given Bluetooth MAC address.
// An optional channelID can be provided; if omitted or 0, default RFCOMM channel 1 is used.
func Connect(address string, channelID ...uint8) (*Client, error) {
	return printer.Connect(address, channelID...)
}

// NewClient creates a new Client instance. If no provider is specified, the platform's
// default Bluetooth transport provider is used.
func NewClient(provider ...transport.Provider) *Client {
	return printer.NewClient(provider...)
}

// ScanDevices performs Bluetooth inquiry for nearby devices.
// Any onFound callbacks provided will be invoked sequentially as devices are discovered.
func ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error) {
	return printer.ScanDevices(onFound...)
}

// GetPairedDevices retrieves Bluetooth devices currently paired with the host system.
func GetPairedDevices() ([]DeviceInfo, error) {
	return printer.GetPairedDevices()
}
