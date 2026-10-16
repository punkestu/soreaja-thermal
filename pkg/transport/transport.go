// Package transport defines abstractions and implementations for printer communication channels.
//
// It encapsulates Bluetooth RFCOMM discovery, device pairing queries, and byte streaming,
// decoupling platform-specific hardware drivers from the core printer and ESC/POS logic.
package transport

import (
	"errors"
	"io"
)

// Standard transport and link lifecycle errors.
var (
	// ErrNotConnected is returned when an I/O operation is attempted without an active connection.
	ErrNotConnected = errors.New("printer is not connected")
	// ErrAlreadyConnected is returned when a connect request is made while already connected.
	ErrAlreadyConnected = errors.New("printer is already connected")
	// ErrConnectionFailed is returned when an RFCOMM or socket connection could not be established.
	ErrConnectionFailed = errors.New("failed to connect to RFCOMM channel")
	// ErrUnsupportedOS is returned when no Bluetooth provider is available for the current runtime OS.
	ErrUnsupportedOS = errors.New("bluetooth transport is not supported on this operating system")
)

// DeviceInfo represents a detected or paired Bluetooth device.
type DeviceInfo struct {
	// Name is the human-readable advertised Bluetooth name of the device.
	Name string
	// Address is the hardware MAC address or UUID formatted string (e.g. "00-11-22-33-44-55").
	Address string
}

// Transport represents an active bidirectional data link to a printer.
type Transport interface {
	io.Writer
	io.Closer
	// IsConnected reports whether the transport channel is open and available for writes.
	IsConnected() bool
}

// Provider represents a platform Bluetooth discovery and connection factory.
type Provider interface {
	// ScanDevices executes an active Bluetooth inquiry for nearby devices.
	// Optional onFound callbacks are called as devices are discovered.
	ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error)
	// GetPairedDevices returns devices already paired with the host system.
	GetPairedDevices() ([]DeviceInfo, error)
	// Connect establishes a transport connection to the specified address on channelID.
	Connect(address string, channelID uint8) (Transport, error)
}
