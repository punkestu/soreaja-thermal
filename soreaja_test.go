package soreaja_test

import (
	"testing"

	soreaja "github.com/punkestu/soreaja-thermal"
	"github.com/punkestu/soreaja-thermal/pkg/transport"
)

func TestRootPackageFacade(t *testing.T) {
	mockProv := transport.NewMockProvider()
	transport.SetDefaultProvider(mockProv)

	devices, err := soreaja.GetPairedDevices()
	if err != nil {
		t.Fatalf("GetPairedDevices error: %v", err)
	}
	if len(devices) == 0 {
		t.Fatalf("Expected mock devices")
	}

	client := soreaja.NewClient(mockProv)
	if err := client.Connect(devices[0].Address, 1); err != nil {
		t.Fatalf("Connect error: %v", err)
	}
	defer client.Disconnect()

	if !client.IsConnected() {
		t.Fatalf("Client should be connected")
	}

	_ = client.SetAlign(soreaja.AlignCenter)
	_ = client.PrintText("Root package test\n")
	_ = client.CutPaper(3)
}
