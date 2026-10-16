//go:build darwin && cgo

package transport

/*
#cgo CFLAGS: -I${SRCDIR}/../../bridge/mac
#cgo LDFLAGS: -L${SRCDIR}/../../bridge/mac -lprinterbridge -Wl,-rpath,${SRCDIR}/../../bridge/mac
#include "bridge.h"
#include <stdlib.h>

extern void goDarwinDeviceCallbackTrampoline(char *name, char *address);

static inline void c_scan_devices() {
    printer_scan_devices((void (*)(const char*, const char*))goDarwinDeviceCallbackTrampoline);
}

static inline void c_get_paired_devices() {
    printer_get_paired_devices((void (*)(const char*, const char*))goDarwinDeviceCallbackTrampoline);
}
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

var (
	darwinMu             sync.Mutex
	activeDeviceCallback func(DeviceInfo)
)

//export goDarwinDeviceCallbackTrampoline
func goDarwinDeviceCallbackTrampoline(cName *C.char, cAddr *C.char) {
	name := "Unknown"
	if cName != nil {
		name = C.GoString(cName)
	}
	addr := "Unknown"
	if cAddr != nil {
		addr = C.GoString(cAddr)
	}

	darwinMu.Lock()
	cb := activeDeviceCallback
	darwinMu.Unlock()

	if cb != nil {
		cb(DeviceInfo{
			Name:    name,
			Address: addr,
		})
	}
}

// DarwinProvider provides Bluetooth discovery and RFCOMM connections on macOS using IOBluetooth.
type DarwinProvider struct {
	mu sync.Mutex
}

func init() {
	defaultProvider = &DarwinProvider{}
}

// ScanDevices performs an active 5-second inquiry for nearby Bluetooth devices via IOBluetooth.
// The optional onFound callback is called on each uniquely discovered device.
func (p *DarwinProvider) ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var devices []DeviceInfo
	seen := make(map[string]bool)

	darwinMu.Lock()
	activeDeviceCallback = func(dev DeviceInfo) {
		if !seen[dev.Address] {
			seen[dev.Address] = true
			devices = append(devices, dev)
			for _, cb := range onFound {
				if cb != nil {
					cb(dev)
				}
			}
		}
	}
	darwinMu.Unlock()

	C.c_scan_devices()

	darwinMu.Lock()
	activeDeviceCallback = nil
	darwinMu.Unlock()

	return devices, nil
}

// GetPairedDevices retrieves paired Bluetooth devices known to macOS IOBluetooth.
func (p *DarwinProvider) GetPairedDevices() ([]DeviceInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var devices []DeviceInfo
	seen := make(map[string]bool)

	darwinMu.Lock()
	activeDeviceCallback = func(dev DeviceInfo) {
		if !seen[dev.Address] {
			seen[dev.Address] = true
			devices = append(devices, dev)
		}
	}
	darwinMu.Unlock()

	C.c_get_paired_devices()

	darwinMu.Lock()
	activeDeviceCallback = nil
	darwinMu.Unlock()

	return devices, nil
}

// Connect establishes an RFCOMM link to a Bluetooth printer at address on channelID.
// If channelID is 0, the printer's first advertised SPP channel or default channel 1 is used.
func (p *DarwinProvider) Connect(address string, channelID uint8) (Transport, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if bool(C.printer_is_connected()) {
		C.printer_disconnect()
	}

	cAddr := C.CString(address)
	defer C.free(unsafe.Pointer(cAddr))

	var ok bool
	if channelID > 0 {
		ok = bool(C.printer_connect_channel(cAddr, C.uint8_t(channelID)))
	} else {
		ok = bool(C.printer_connect(cAddr))
	}

	if !ok {
		return nil, fmt.Errorf("%w: %s on channel %d", ErrConnectionFailed, address, channelID)
	}

	return &darwinTransport{
		address:   address,
		channelID: channelID,
	}, nil
}

type darwinTransport struct {
	address   string
	channelID uint8
	mu        sync.Mutex
	closed    bool
}

func (t *darwinTransport) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed || !bool(C.printer_is_connected()) {
		return 0, ErrNotConnected
	}

	if len(p) == 0 {
		return 0, nil
	}

	ok := bool(C.printer_send_raw((*C.uint8_t)(unsafe.Pointer(&p[0])), C.int32_t(len(p))))
	if !ok {
		return 0, fmt.Errorf("failed to transmit %d bytes over RFCOMM", len(p))
	}

	return len(p), nil
}

func (t *darwinTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.closed {
		C.printer_disconnect()
		t.closed = true
	}
	return nil
}

func (t *darwinTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return false
	}
	return bool(C.printer_is_connected())
}
