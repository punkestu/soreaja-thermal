// Package soreaja provides a high-level Go client and ESC/POS thermal printing engine.
package soreaja

import (
	"github.com/punkestu/soreaja-thermal/pkg/escpos"
	"github.com/punkestu/soreaja-thermal/pkg/printer"
	"github.com/punkestu/soreaja-thermal/pkg/transport"
)

// Re-exported types
type Client = printer.Client
type DeviceInfo = transport.DeviceInfo
type Alignment = escpos.Alignment

// Text alignment constants
const (
	AlignLeft   = escpos.AlignLeft
	AlignCenter = escpos.AlignCenter
	AlignRight  = escpos.AlignRight
)

// Common errors
var (
	ErrNotConnected     = transport.ErrNotConnected
	ErrAlreadyConnected = transport.ErrAlreadyConnected
	ErrConnectionFailed = transport.ErrConnectionFailed
)

// Re-exported functions
var (
	Connect          = printer.Connect
	NewClient        = printer.NewClient
	ScanDevices      = printer.ScanDevices
	GetPairedDevices = printer.GetPairedDevices
)
