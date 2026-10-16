package escpos

import (
	"image"
	"image/color"
	"io"
	"os"

	_ "image/jpeg"
	_ "image/png"
)

// RasterizeFile reads a PNG or JPEG image file from disk and converts it into
// ESC/POS raster chunks (GS v 0 0) formatted for thermal receipt printers.
// maxWidth specifies the target width in dots (typically 384 for 58mm paper or 576 for 80mm paper).
// If dither is true, Floyd-Steinberg error diffusion is applied; otherwise a 50% luminance threshold is used.
func RasterizeFile(path string, maxWidth int, dither bool) ([][]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return RasterizeReader(file, maxWidth, dither)
}

// RasterizeReader decodes an image byte stream from an io.Reader and converts it into ESC/POS raster chunks.
// It auto-scales the image if wider than maxWidth while preserving the aspect ratio.
func RasterizeReader(r io.Reader, maxWidth int, dither bool) ([][]byte, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return RasterizeImage(img, maxWidth, dither, 128), nil
}

// RasterizeImage converts any image.Image into sliced chunks of ESC/POS GS v 0 0 raster commands.
//
// Parameters:
//   - src: The decoded source image (PNG, JPEG, etc.).
//   - maxWidth: Maximum dot width (usually 384 for 58mm printer, 576 for 80mm printer; <= 0 keeps source width).
//   - dither: If true, applies Floyd-Steinberg error diffusion for smooth gradient reproduction; if false, uses 50% luminance threshold.
//   - chunkHeight: Maximum vertical dot rows per ESC/POS command slice (default 128) to prevent printer hardware buffer overflow.
//
// Returns a slice of byte buffers, each containing an individual ESC/POS GS v 0 command block.
func RasterizeImage(src image.Image, maxWidth int, dither bool, chunkHeight int) [][]byte {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == 0 || srcH == 0 {
		return nil
	}

	targetW := srcW
	targetH := srcH

	if maxWidth > 0 && targetW > maxWidth {
		scale := float64(maxWidth) / float64(targetW)
		targetW = maxWidth
		targetH = int(float64(targetH) * scale)
	}

	if targetW <= 0 || targetH <= 0 {
		return nil
	}

	// 1. Resample / compute grayscale matrix with alpha blending over white
	gray := make([]float32, targetW*targetH)
	for y := 0; y < targetH; y++ {
		srcY := bounds.Min.Y + (y * srcH / targetH)
		for x := 0; x < targetW; x++ {
			srcX := bounds.Min.X + (x * srcW / targetW)
			c := src.At(srcX, srcY)
			r, g, b, a := c.RGBA() // 0 to 65535

			rf := float32(r >> 8)
			gf := float32(g >> 8)
			bf := float32(b >> 8)
			af := float32(a>>8) / 255.0

			// Luminance formula (ITU-R BT.601)
			lum := 0.299*rf + 0.587*gf + 0.114*bf
			// Blend over white background
			blended := lum*af + 255.0*(1.0-af)
			gray[y*targetW+x] = blended
		}
	}

	// 2. Monochrome 1-bit dot matrix (true = black/printed, false = white)
	mono := make([]bool, targetW*targetH)

	if dither {
		// Floyd-Steinberg error diffusion
		for y := 0; y < targetH; y++ {
			for x := 0; x < targetW; x++ {
				idx := y*targetW + x
				oldVal := gray[idx]
				var newVal float32 = 255.0
				if oldVal < 128.0 {
					newVal = 0.0
				}
				mono[idx] = (newVal == 0.0)
				err := oldVal - newVal

				if x+1 < targetW {
					gray[y*targetW+(x+1)] += err * (7.0 / 16.0)
				}
				if y+1 < targetH {
					if x > 0 {
						gray[(y+1)*targetW+(x-1)] += err * (3.0 / 16.0)
					}
					gray[(y+1)*targetW+x] += err * (5.0 / 16.0)
					if x+1 < targetW {
						gray[(y+1)*targetW+(x+1)] += err * (1.0 / 16.0)
					}
				}
			}
		}
	} else {
		// Simple 50% luminance thresholding
		for i := 0; i < targetW*targetH; i++ {
			mono[i] = gray[i] < 128.0
		}
	}

	// 3. Encode into ESC/POS GS v 0 0 raster chunks: GS v 0 0 xL xH yL yH [data...]
	if chunkHeight <= 0 {
		chunkHeight = 128
	}

	widthBytes := (targetW + 7) / 8
	var chunks [][]byte
	currentY := 0

	for currentY < targetH {
		sliceH := chunkHeight
		if currentY+sliceH > targetH {
			sliceH = targetH - currentY
		}

		xL := byte(widthBytes & 0xFF)
		xH := byte((widthBytes >> 8) & 0xFF)
		yL := byte(sliceH & 0xFF)
		yH := byte((sliceH >> 8) & 0xFF)

		// Header: GS v 0 0 xL xH yL yH
		sliceBytes := make([]byte, 0, 8+sliceH*widthBytes)
		sliceBytes = append(sliceBytes, 0x1D, 0x76, 0x30, 0x00, xL, xH, yL, yH)

		for row := 0; row < sliceH; row++ {
			actualY := currentY + row
			for byteCol := 0; byteCol < widthBytes; byteCol++ {
				var b byte
				for bit := 0; bit < 8; bit++ {
					actualX := byteCol*8 + bit
					if actualX < targetW && mono[actualY*targetW+actualX] {
						b |= 1 << (7 - bit)
					}
				}
				sliceBytes = append(sliceBytes, b)
			}
		}

		chunks = append(chunks, sliceBytes)
		currentY += sliceH
	}

	return chunks
}

// Ensure color is imported
var _ color.Color = nil
