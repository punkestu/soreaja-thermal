package escpos

// QRCode generates ESC/POS Model 2 QR code sequence commands (GS ( k ...) for the provided text.
//
// Parameters:
//   - text: The string content or URL to encode in the QR code.
//   - size: Module unit size in dots (range 1-16; typical values are 4 to 8 for thermal receipts).
//
// The error correction level is configured to M (approximately 15% recovery capacity).
func QRCode(text string, size uint8) []byte {
	if len(text) == 0 {
		return nil
	}

	data := []byte(text)
	var bytes []byte

	// 1. Select Model: Model 2 (0x32)
	bytes = append(bytes, 0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00)

	// 2. Set Module Size (1-16)
	if size < 1 {
		size = 1
	} else if size > 16 {
		size = 16
	}
	bytes = append(bytes, 0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, size)

	// 3. Set Error Correction Level: M (15% recovery, 0x31)
	bytes = append(bytes, 0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x45, 0x31)

	// 4. Store QR Data: length = len(data) + 3
	length := len(data) + 3
	pL := byte(length & 0xFF)
	pH := byte((length >> 8) & 0xFF)
	bytes = append(bytes, 0x1D, 0x28, 0x6B, pL, pH, 0x31, 0x50, 0x30)
	bytes = append(bytes, data...)

	// 5. Print Stored QR Code
	bytes = append(bytes, 0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30)

	return bytes
}
