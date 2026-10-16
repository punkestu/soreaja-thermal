package escpos

// Alignment defines text horizontal alignment.
type Alignment uint8

const (
	AlignLeft   Alignment = 0
	AlignCenter Alignment = 1
	AlignRight  Alignment = 2
)

// Init resets the printer to its default state.
func Init() []byte {
	return []byte{0x1B, 0x40} // ESC @
}

// LineFeed returns a single line feed byte.
func LineFeed() []byte {
	return []byte{0x0A} // LF
}

// Feed returns the ESC/POS command to feed n lines.
func Feed(lines uint8) []byte {
	return []byte{0x1B, 0x64, lines} // ESC d n
}

// Cut returns the ESC/POS command to feed specified lines and perform a partial/full cut.
func Cut(feedLines uint8) []byte {
	var out []byte
	if feedLines > 0 {
		out = append(out, 0x1B, 0x64, feedLines)
	}
	out = append(out, 0x1D, 0x56, 0x42, 0x00) // GS V 'B' 0
	return out
}

// Align sets text alignment: Left (0), Center (1), Right (2).
func Align(alignment Alignment) []byte {
	return []byte{0x1B, 0x61, byte(alignment)} // ESC a n
}

// Bold enables or disables bold text emphasis.
func Bold(enable bool) []byte {
	val := byte(0)
	if enable {
		val = 1
	}
	return []byte{0x1B, 0x45, val} // ESC E n
}

// Underline enables or disables single-line underline.
func Underline(enable bool) []byte {
	val := byte(0)
	if enable {
		val = 1
	}
	return []byte{0x1B, 0x2D, val} // ESC - n
}

// Invert enables or disables reverse (white on black) printing mode.
func Invert(enable bool) []byte {
	val := byte(0)
	if enable {
		val = 1
	}
	return []byte{0x1D, 0x42, val} // GS B n
}

// FontSize sets character width and height scaling (0 to 7 multiplier each, where 0 is 1x and 7 is 8x).
func FontSize(width, height uint8) []byte {
	if width > 7 {
		width = 7
	}
	if height > 7 {
		height = 7
	}
	n := (width << 4) | height
	return []byte{0x1D, 0x21, n} // GS ! n
}

// CashDrawer generates pulse to open standard cash drawer connected to RJ11 port.
func CashDrawer() []byte {
	return []byte{0x1B, 0x70, 0x00, 0x19, 0xFA} // ESC p 0 25 250
}
