package printer

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/punkestu/soreaja-thermal/pkg/escpos"
	"github.com/punkestu/soreaja-thermal/pkg/transport"
)

// Re-export common types for caller convenience
type Alignment = escpos.Alignment
type DeviceInfo = transport.DeviceInfo

const (
	AlignLeft   = escpos.AlignLeft
	AlignCenter = escpos.AlignCenter
	AlignRight  = escpos.AlignRight
)

var (
	ErrNotConnected    = transport.ErrNotConnected
	ErrAlreadyConnected = transport.ErrAlreadyConnected
	ErrConnectionFailed = transport.ErrConnectionFailed
	ErrInvalidMAC      = errors.New("invalid bluetooth MAC address")
)

// ScanDevices performs Bluetooth inquiry for nearby devices.
func ScanDevices(onFound ...func(DeviceInfo)) ([]DeviceInfo, error) {
	prov, err := transport.DefaultProvider()
	if err != nil {
		return nil, err
	}
	return prov.ScanDevices(onFound...)
}

// GetPairedDevices retrieves paired Bluetooth devices.
func GetPairedDevices() ([]DeviceInfo, error) {
	prov, err := transport.DefaultProvider()
	if err != nil {
		return nil, err
	}
	return prov.GetPairedDevices()
}

// Connect establishes a connection to a thermal printer at the given address.
func Connect(address string, channelID ...uint8) (*Client, error) {
	prov, err := transport.DefaultProvider()
	if err != nil {
		return nil, err
	}
	c := NewClient(prov)
	ch := uint8(1)
	if len(channelID) > 0 && channelID[0] > 0 {
		ch = channelID[0]
	}
	if err := c.Connect(address, ch); err != nil {
		return nil, err
	}
	return c, nil
}

// Client manages communication and command formatting for a thermal printer.
type Client struct {
	provider transport.Provider
	conn     transport.Transport
	address  string
	channel  uint8
	mu       sync.Mutex
}

// NewClient creates a new Client instance. If no provider is given, DefaultProvider is used.
func NewClient(provider ...transport.Provider) *Client {
	var p transport.Provider
	if len(provider) > 0 && provider[0] != nil {
		p = provider[0]
	} else {
		p, _ = transport.DefaultProvider()
	}
	return &Client{provider: p}
}

// Connect opens an RFCOMM channel to the target device address.
func (c *Client) Connect(address string, channelID uint8) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.provider == nil {
		return transport.ErrUnsupportedOS
	}

	if c.conn != nil && c.conn.IsConnected() {
		_ = c.conn.Close()
	}

	conn, err := c.provider.Connect(address, channelID)
	if err != nil {
		return err
	}

	c.conn = conn
	c.address = address
	c.channel = channelID
	return nil
}

// Disconnect closes the current connection.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// IsConnected returns whether the link is currently open.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && c.conn.IsConnected()
}

// SendRaw transmits raw bytes to the printer.
func (c *Client) SendRaw(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || !c.conn.IsConnected() {
		return ErrNotConnected
	}
	_, err := c.conn.Write(data)
	return err
}

// Init resets the printer to default configuration.
func (c *Client) Init() error {
	return c.SendRaw(escpos.Init())
}

// LineFeed sends a line feed character.
func (c *Client) LineFeed() error {
	return c.SendRaw(escpos.LineFeed())
}

// Feed feeds n lines of paper.
func (c *Client) Feed(lines uint8) error {
	return c.SendRaw(escpos.Feed(lines))
}

// CutPaper cuts the paper after feeding specified lines.
func (c *Client) CutPaper(feedLines uint8) error {
	return c.SendRaw(escpos.Cut(feedLines))
}

// SetAlign sets horizontal text alignment.
func (c *Client) SetAlign(align Alignment) error {
	return c.SendRaw(escpos.Align(align))
}

// SetBold toggles bold text emphasis.
func (c *Client) SetBold(enable bool) error {
	return c.SendRaw(escpos.Bold(enable))
}

// SetUnderline toggles single-underline.
func (c *Client) SetUnderline(enable bool) error {
	return c.SendRaw(escpos.Underline(enable))
}

// SetFontSize sets character font scale multipliers (0-7 width, 0-7 height).
func (c *Client) SetFontSize(width, height uint8) error {
	return c.SendRaw(escpos.FontSize(width, height))
}

// PrintText writes a string of text directly to the printer.
func (c *Client) PrintText(text string) error {
	return c.SendRaw([]byte(text))
}

// PrintLine writes text followed by a newline.
func (c *Client) PrintLine(text string) error {
	return c.SendRaw([]byte(text + "\n"))
}

// PrintQR generates and prints an ESC/POS Model 2 QR code.
func (c *Client) PrintQR(text string, size uint8) error {
	qrBytes := escpos.QRCode(text, size)
	if len(qrBytes) == 0 {
		return errors.New("empty QR content")
	}
	return c.SendRaw(qrBytes)
}

// PrintImage rasterizes an image file from disk and transmits it in chunks.
func (c *Client) PrintImage(filePath string, maxWidth int, dither bool) error {
	chunks, err := escpos.RasterizeFile(filePath, maxWidth, dither)
	if err != nil {
		return fmt.Errorf("failed to rasterize image: %w", err)
	}
	if len(chunks) == 0 {
		return errors.New("no image data generated")
	}

	// Center align image
	_ = c.SetAlign(AlignCenter)

	for _, chunk := range chunks {
		if err := c.SendRaw(chunk); err != nil {
			return err
		}
		// Small transmission pacing delay for thermal head
		time.Sleep(20 * time.Millisecond)
	}

	_ = c.Feed(2)
	return nil
}

// PrintTestReceipt prints a comprehensive diagnostic receipt.
func (c *Client) PrintTestReceipt(widthDots int) error {
	var buf []byte

	// 1. Initialize
	buf = append(buf, escpos.Init()...)

	// 2. Header
	buf = append(buf, escpos.Align(AlignCenter)...)
	buf = append(buf, escpos.FontSize(1, 1)...)
	buf = append(buf, escpos.Bold(true)...)
	buf = append(buf, []byte("SOREAJA THERMAL\n")...)
	buf = append(buf, escpos.FontSize(0, 0)...)
	buf = append(buf, escpos.Bold(false)...)
	buf = append(buf, []byte("Portable ESC/POS Go Bridge\n")...)
	buf = append(buf, []byte("--------------------------------\n")...)

	// 3. System info
	buf = append(buf, escpos.Align(AlignLeft)...)
	buf = append(buf, []byte(fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05")))...)
	buf = append(buf, []byte(fmt.Sprintf("Paper Dots: %d dots\n", widthDots))...)
	buf = append(buf, []byte("Driver: Pure Go + Swift Bridge\n")...)
	buf = append(buf, []byte("--------------------------------\n")...)

	// 4. Formatting styles test
	buf = append(buf, escpos.Bold(true)...)
	buf = append(buf, []byte("Bold text test\n")...)
	buf = append(buf, escpos.Bold(false)...)

	buf = append(buf, escpos.Underline(true)...)
	buf = append(buf, []byte("Underlined text test\n")...)
	buf = append(buf, escpos.Underline(false)...)

	buf = append(buf, escpos.FontSize(1, 0)...)
	buf = append(buf, []byte("Double Width\n")...)
	buf = append(buf, escpos.FontSize(0, 1)...)
	buf = append(buf, []byte("Double Height\n")...)
	buf = append(buf, escpos.FontSize(0, 0)...)

	// 5. Alignment test
	buf = append(buf, escpos.Align(AlignLeft)...)
	buf = append(buf, []byte("Left aligned\n")...)
	buf = append(buf, escpos.Align(AlignCenter)...)
	buf = append(buf, []byte("Center aligned\n")...)
	buf = append(buf, escpos.Align(AlignRight)...)
	buf = append(buf, []byte("Right aligned\n")...)
	buf = append(buf, []byte("--------------------------------\n")...)

	// 6. QR Code
	buf = append(buf, escpos.Align(AlignCenter)...)
	buf = append(buf, []byte("Scan to verify QR:\n")...)
	buf = append(buf, escpos.QRCode("https://github.com/punkestu/soreaja-thermal", 6)...)
	buf = append(buf, []byte("\nAll diagnostics passed!\n")...)
	buf = append(buf, []byte("================================\n")...)

	// 7. Cut paper
	buf = append(buf, escpos.Cut(4)...)

	return c.SendRaw(buf)
}
