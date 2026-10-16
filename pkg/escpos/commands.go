// Package escpos implements the ESC/POS thermal printer command protocol and image rasterization.
//
// It provides byte-level generation for text styling, alignment, cuts, cash drawer kickouts,
// Model 2 QR codes, and Floyd-Steinberg error diffusion raster bitmaps.
package escpos

// Alignment defines horizontal text alignment for ESC/POS commands.
type Alignment uint8

const (
	// AlignLeft sets left-justified text mode (ESC a 0).
	AlignLeft Alignment = 0
	// AlignCenter sets centered text mode (ESC a 1).
	AlignCenter Alignment = 1
	// AlignRight sets right-justified text mode (ESC a 2).
	AlignRight Alignment = 2
)

// Init returns the ESC/POS initialize printer command (ESC @: 0x1B, 0x40).
// It resets the printer hardware to its default power-on state, clearing line buffers
// and resetting print styles.
func Init() []byte {
	return []byte{0x1B, 0x40} // ESC @
}

// LineFeed returns a single line feed control byte (LF: 0x0A).
func LineFeed() []byte {
	return []byte{0x0A} // LF
}

// Feed returns the ESC/POS command to print and feed paper forward by n lines (ESC d n).
func Feed(lines uint8) []byte {
	return []byte{0x1B, 0x64, lines} // ESC d n
}

// Cut returns the ESC/POS command to feed specified lines and perform a partial or full cut (GS V B n).
func Cut(feedLines uint8) []byte {
	var out []byte
	if feedLines > 0 {
		out = append(out, 0x1B, 0x64, feedLines)
	}
	out = append(out, 0x1D, 0x56, 0x42, 0x00) // GS V 'B' 0
	return out
}

// Align returns the ESC/POS command to set horizontal text alignment (ESC a n).
// Valid values are AlignLeft (0), AlignCenter (1), or AlignRight (2).
func Align(alignment Alignment) []byte {
	return []byte{0x1B, 0x61, byte(alignment)} // ESC a n
}

// Bold returns the ESC/POS command to enable or disable emphasized/bold printing (ESC E n).
func Bold(enable bool) []byte {
	val := byte(0)
	if enable {
		val = 1
	}
	return []byte{0x1B, 0x45, val} // ESC E n
}

// Underline returns the ESC/POS command to enable or disable single-line underline (ESC - n).
func Underline(enable bool) []byte {
	val := byte(0)
	if enable {
		val = 1
	}
	return []byte{0x1B, 0x2D, val} // ESC - n
}

// Invert returns the ESC/POS command to enable or disable reverse (white-on-black) printing (GS B n).
func Invert(enable bool) []byte {
	val := byte(0)
	if enable {
		val = 1
	}
	return []byte{0x1D, 0x42, val} // GS B n
}

// FontSize returns the ESC/POS command to set character width and height scaling (GS ! n).
// width and height are multipliers from 0 to 7 (where 0 corresponds to normal 1x and 7 corresponds to 8x).
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

// CashDrawer returns the ESC/POS pulse command to open a standard RJ11 cash drawer (ESC p 0 25 250).
func CashDrawer() []byte {
	return []byte{0x1B, 0x70, 0x00, 0x19, 0xFA} // ESC p 0 25 250
}
