package transport_test

import (
	"bytes"
	"testing"

	"github.com/punkestu/soreaja-thermal/pkg/transport"
)

func TestMockTransport(t *testing.T) {
	mockProv := transport.NewMockProvider()
	transport.SetDefaultProvider(mockProv)

	prov, err := transport.DefaultProvider()
	if err != nil {
		t.Fatalf("DefaultProvider error: %v", err)
	}

	paired, err := prov.GetPairedDevices()
	if err != nil {
		t.Fatalf("GetPairedDevices error: %v", err)
	}
	if len(paired) != 1 || paired[0].Address != "00:11:22:33:44:55" {
		t.Fatalf("Unexpected paired devices: %+v", paired)
	}

	conn, err := prov.Connect("00:11:22:33:44:55", 1)
	if err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer conn.Close()

	if !conn.IsConnected() {
		t.Errorf("Expected transport to be connected")
	}

	payload := []byte("Hello Thermal Printer\n")
	n, err := conn.Write(payload)
	if err != nil || n != len(payload) {
		t.Fatalf("Write failed: n=%d err=%v", n, err)
	}

	mockTrans := conn.(*transport.MockTransport)
	if !bytes.Equal(mockTrans.Buffer.Bytes(), payload) {
		t.Errorf("Buffer content mismatch: got %q, want %q", mockTrans.Buffer.String(), payload)
	}

	_ = conn.Close()
	if conn.IsConnected() {
		t.Errorf("Expected transport to be disconnected after Close()")
	}

	_, err = conn.Write([]byte("more data"))
	if err != transport.ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
}
