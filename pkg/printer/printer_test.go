package printer_test

import (
	"bytes"
	"testing"

	"github.com/punkestu/soreaja-thermal/pkg/escpos"
	"github.com/punkestu/soreaja-thermal/pkg/printer"
	"github.com/punkestu/soreaja-thermal/pkg/transport"
)

func TestPrinterClientWithMock(t *testing.T) {
	mockProv := transport.NewMockProvider()
	client := printer.NewClient(mockProv)

	if client.IsConnected() {
		t.Fatalf("Client should not be connected initially")
	}

	err := client.Connect("00:11:22:33:44:55", 1)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Disconnect()

	if !client.IsConnected() {
		t.Fatalf("Client should be connected")
	}

	// Test Init
	if err := client.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test PrintText
	if err := client.PrintText("Hello World"); err != nil {
		t.Fatalf("PrintText failed: %v", err)
	}

	// Test Formatting
	if err := client.SetAlign(printer.AlignCenter); err != nil {
		t.Fatalf("SetAlign failed: %v", err)
	}
	if err := client.SetBold(true); err != nil {
		t.Fatalf("SetBold failed: %v", err)
	}
	if err := client.PrintLine("Centered Bold"); err != nil {
		t.Fatalf("PrintLine failed: %v", err)
	}

	// Test QR Code
	if err := client.PrintQR("https://example.com", 6); err != nil {
		t.Fatalf("PrintQR failed: %v", err)
	}

	// Test Cut
	if err := client.CutPaper(3); err != nil {
		t.Fatalf("CutPaper failed: %v", err)
	}

	mockTrans := mockProv.CurrentTrans
	data := mockTrans.Buffer.Bytes()
	if len(data) == 0 {
		t.Fatalf("No data was written to mock transport")
	}

	// Verify buffer contains init
	if !bytes.Contains(data, escpos.Init()) {
		t.Errorf("Buffer missing init command")
	}
	// Verify buffer contains text
	if !bytes.Contains(data, []byte("Hello World")) {
		t.Errorf("Buffer missing text")
	}
	// Verify buffer contains cut command
	if !bytes.Contains(data, escpos.Cut(3)) {
		t.Errorf("Buffer missing cut command")
	}

	// Test Disconnect
	if err := client.Disconnect(); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}
	if client.IsConnected() {
		t.Fatalf("Client should be disconnected")
	}
}
