package transport

import (
	"errors"
	"io"
)

var (
	ErrNotConnected    = errors.New("printer is not connected")
	ErrAlreadyConnected = errors.New("printer is already connected")
	ErrConnectionFailed = errors.New("failed to connect to RFCOMM channel")
	ErrUnsupportedOS   = errors.New("bluetooth transport is not supported on this operating system")
)

// DeviceInfo represents a detected or paired Bluetooth device.
type DeviceInfo struct {
	Name    string
	Address string
}

// Transport represents an active data link to a printer.
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
