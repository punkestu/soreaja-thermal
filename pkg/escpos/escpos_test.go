package escpos_test

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/punkestu/soreaja-thermal/pkg/escpos"
)

func TestCommands(t *testing.T) {
	// Test Init
	if !bytes.Equal(escpos.Init(), []byte{0x1B, 0x40}) {
		t.Errorf("Init failed, got %v", escpos.Init())
	}

	// Test LineFeed
	if !bytes.Equal(escpos.LineFeed(), []byte{0x0A}) {
		t.Errorf("LineFeed failed, got %v", escpos.LineFeed())
	}

	// Test Feed
	if !bytes.Equal(escpos.Feed(3), []byte{0x1B, 0x64, 0x03}) {
		t.Errorf("Feed failed, got %v", escpos.Feed(3))
	}

	// Test Align
	if !bytes.Equal(escpos.Align(escpos.AlignCenter), []byte{0x1B, 0x61, 0x01}) {
		t.Errorf("Align failed, got %v", escpos.Align(escpos.AlignCenter))
	}

	// Test Bold
	if !bytes.Equal(escpos.Bold(true), []byte{0x1B, 0x45, 0x01}) {
		t.Errorf("Bold true failed, got %v", escpos.Bold(true))
	}
	if !bytes.Equal(escpos.Bold(false), []byte{0x1B, 0x45, 0x00}) {
		t.Errorf("Bold false failed, got %v", escpos.Bold(false))
	}

	// Test Cut
	cut := escpos.Cut(3)
	expectedCut := []byte{0x1B, 0x64, 0x03, 0x1D, 0x56, 0x42, 0x00}
	if !bytes.Equal(cut, expectedCut) {
		t.Errorf("Cut failed, got %v, want %v", cut, expectedCut)
	}

	// Test FontSize
	fs := escpos.FontSize(1, 2)
	expectedFS := []byte{0x1D, 0x21, (1 << 4) | 2}
	if !bytes.Equal(fs, expectedFS) {
		t.Errorf("FontSize failed, got %v, want %v", fs, expectedFS)
	}
}

func TestQRCode(t *testing.T) {
	qr := escpos.QRCode("https://example.com", 6)
	if len(qr) == 0 {
		t.Fatalf("QRCode returned empty bytes")
	}

	// Check model 2 prefix: 1D 28 6B 04 00 31 41 32 00
	model2Prefix := []byte{0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00}
	if !bytes.HasPrefix(qr, model2Prefix) {
		t.Errorf("QRCode does not start with Model 2 header")
	}

	// Check print suffix: 1D 28 6B 03 00 31 51 30
	printSuffix := []byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30}
	if !bytes.HasSuffix(qr, printSuffix) {
		t.Errorf("QRCode does not end with print command")
	}
}

func TestRasterizeImage(t *testing.T) {
	// Create an 8x8 checkerboard image
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	chunks := escpos.RasterizeImage(img, 8, false, 8)
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	// Header: 1D 76 30 00 xL xH yL yH -> 1D 76 30 00 01 00 08 00
	expectedHeader := []byte{0x1D, 0x76, 0x30, 0x00, 0x01, 0x00, 0x08, 0x00}
	if !bytes.HasPrefix(chunks[0], expectedHeader) {
		t.Errorf("Raster chunk header mismatch, got %v", chunks[0][:8])
	}

	// Data rows: (10101010b = 0xAA, 01010101b = 0x55)
	data := chunks[0][8:]
	if len(data) != 8 {
		t.Fatalf("Expected 8 bytes data, got %d", len(data))
	}
	if data[0] != 0xAA || data[1] != 0x55 {
		t.Errorf("Pattern mismatch: row0=%x row1=%x", data[0], data[1])
	}
}
