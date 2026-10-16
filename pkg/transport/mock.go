package transport

import (
	"bytes"
	"sync"
)

// MockTransport records written bytes into a buffer in memory.
type MockTransport struct {
	mu        sync.Mutex
	Buffer    bytes.Buffer
	connected bool
}

func NewMockTransport() *MockTransport {
	return &MockTransport{connected: true}
}

func (m *MockTransport) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return 0, ErrNotConnected
	}
	return m.Buffer.Write(p)
}

func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

func (m *MockTransport) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

// MockProvider provides mock devices and creates MockTransport instances.
type MockProvider struct {
	mu           sync.Mutex
	Devices      []DeviceInfo
	LastChannel  uint8
	LastAddress  string
	CurrentTrans *MockTransport
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		Devices: []DeviceInfo{
			{Name: "Mock Thermal 58mm", Address: "00:11:22:33:44:55"},
		},
	}
}

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

func (m *MockProvider) GetPairedDevices() ([]DeviceInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Devices, nil
}

func (m *MockProvider) Connect(address string, channelID uint8) (Transport, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastAddress = address
	m.LastChannel = channelID
	m.CurrentTrans = NewMockTransport()
	return m.CurrentTrans, nil
}
