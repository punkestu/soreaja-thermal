package transport

import (
	"bytes"
	"sync"
)

// MockTransport records written bytes into an in-memory buffer without requiring physical hardware.
// It is intended for automated unit tests and continuous integration environments.
type MockTransport struct {
	mu        sync.Mutex
	Buffer    bytes.Buffer
	connected bool
}

// NewMockTransport creates and returns a connected MockTransport instance.
func NewMockTransport() *MockTransport {
	return &MockTransport{connected: true}
}

// Write writes bytes into the internal in-memory buffer, returning ErrNotConnected if closed.
func (m *MockTransport) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return 0, ErrNotConnected
	}
	return m.Buffer.Write(p)
}

// Close marks the mock transport as disconnected.
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

// IsConnected returns whether the mock transport is currently marked connected.
func (m *MockTransport) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

// MockProvider implements the Provider interface with simulated in-memory Bluetooth devices.
type MockProvider struct {
	mu           sync.Mutex
	Devices      []DeviceInfo
	LastChannel  uint8
	LastAddress  string
	CurrentTrans *MockTransport
}

// NewMockProvider creates a new MockProvider initialized with a default mock printer device.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		Devices: []DeviceInfo{
			{Name: "Mock Thermal 58mm", Address: "00:11:22:33:44:55"},
		},
	}
}

// ScanDevices invokes onFound for each mock device and returns the mock device list.
func (m *MockProvider) ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, dev := range m.Devices {
		for _, cb := range onFound {
			if cb != nil {
				cb(dev)
			}
		}
	}
	return m.Devices, nil
}

// GetPairedDevices returns the configured list of mock paired devices.
func (m *MockProvider) GetPairedDevices() ([]DeviceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Devices, nil
}

// Connect simulates opening an RFCOMM link by instantiating and storing a new MockTransport.
func (m *MockProvider) Connect(address string, channelID uint8) (Transport, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastAddress = address
	m.LastChannel = channelID
	m.CurrentTrans = NewMockTransport()
	return m.CurrentTrans, nil
}
