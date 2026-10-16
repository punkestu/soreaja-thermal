package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/punkestu/soreaja-thermal/pkg/escpos"
	"github.com/punkestu/soreaja-thermal/pkg/printer"
)

type UIApp struct {
	window   fyne.Window
	client   *printer.Client
	mu       sync.Mutex
	devices  []printer.DeviceInfo
	devMap   map[string]printer.DeviceInfo // name -> DeviceInfo

	// UI controls
	deviceSelect   *widget.Select
	channelEntry   *widget.Entry
	statusLabel    *widget.Label
	connectBtn     *widget.Button
	widthSelect    *widget.Select
	logText        *widget.Entry
	logScroll      *container.Scroll

	// Text Tab
	textEntry      *widget.Entry
	alignSelect    *widget.RadioGroup
	boldCheck      *widget.Check
	underlineCheck *widget.Check
	fontSizeSelect *widget.Select
	feedCountEntry *widget.Entry

	// Image Tab
	imagePathEntry *widget.Entry
	imagePreview   *canvas.Image
	ditherCheck    *widget.Check

	// QR Tab
	qrTextEntry    *widget.Entry
	qrSizeSlider   *widget.Slider
	qrSizeLabel    *widget.Label

	// Raw Tab
	rawHexEntry    *widget.Entry
}

func main() {
	a := app.NewWithID("com.microapes.soreaja.thermal")
	w := a.NewWindow("Soreaja Thermal Printer — ESC/POS Desktop")
	w.Resize(fyne.NewSize(820, 680))

	ui := &UIApp{
		window:  w,
		client:  printer.NewClient(),
		devMap:  make(map[string]printer.DeviceInfo),
	}

	content := ui.buildUI()
	w.SetContent(content)

	// Initial population of paired devices
	ui.refreshPairedDevices()

	w.ShowAndRun()
}

func (u *UIApp) appendLog(msg string) {
	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] %s\n", ts, msg)
	u.logText.SetText(u.logText.Text + line)
	u.logText.CursorRow = len(strings.Split(u.logText.Text, "\n"))
	if u.logScroll != nil {
		u.logScroll.ScrollToBottom()
	}
}

func (u *UIApp) getPaperWidth() int {
	if u.widthSelect.Selected == "80mm (576 dots)" {
		return 576
	}
	return 384 // default 58mm
}

func (u *UIApp) refreshPairedDevices() {
	u.appendLog("Fetching paired Bluetooth devices...")
	devices, err := printer.GetPairedDevices()
	if err != nil {
		u.appendLog(fmt.Sprintf("Error fetching paired devices: %v", err))
		return
	}

	u.devices = devices
	u.devMap = make(map[string]printer.DeviceInfo)
	var names []string

	for _, d := range devices {
		label := fmt.Sprintf("%s (%s)", d.Name, d.Address)
		names = append(names, label)
		u.devMap[label] = d
	}

	u.deviceSelect.Options = names
	if len(names) > 0 {
		u.deviceSelect.SetSelected(names[0])
		u.appendLog(fmt.Sprintf("Found %d paired device(s).", len(names)))
	} else {
		u.deviceSelect.ClearSelected()
		u.appendLog("No paired devices found. Pair printer in macOS Settings -> Bluetooth.")
	}
}

func (u *UIApp) startScan() {
	u.appendLog("Scanning for Bluetooth devices (5s inquiry)...")
	go func() {
		devices, err := printer.ScanDevices(func(dev printer.DeviceInfo) {
			u.appendLog(fmt.Sprintf("Discovered: %s (%s)", dev.Name, dev.Address))
		})
		if err != nil {
			u.appendLog(fmt.Sprintf("Scan error: %v", err))
			return
		}

		var names []string
		u.devMap = make(map[string]printer.DeviceInfo)
		for _, d := range devices {
			label := fmt.Sprintf("%s (%s)", d.Name, d.Address)
			names = append(names, label)
			u.devMap[label] = d
		}

		u.deviceSelect.Options = names
		if len(names) > 0 && u.deviceSelect.Selected == "" {
			u.deviceSelect.SetSelected(names[0])
		}
		u.appendLog(fmt.Sprintf("Scan complete. Discovered %d device(s).", len(devices)))
	}()
}

func (u *UIApp) toggleConnect() {
	if u.client.IsConnected() {
		u.appendLog("Disconnecting...")
		_ = u.client.Disconnect()
		u.statusLabel.SetText("Status: Disconnected")
		u.connectBtn.SetText("Connect")
		u.connectBtn.Importance = widget.HighImportance
		u.appendLog("Disconnected from printer.")
		return
	}

	selected := u.deviceSelect.Selected
	dev, ok := u.devMap[selected]
	if !ok || dev.Address == "" {
		dialog.ShowInformation("No Device Selected", "Please select a Bluetooth printer from the dropdown first.", u.window)
		return
	}

	ch := 1
	if c, err := strconv.Atoi(u.channelEntry.Text); err == nil && c > 0 {
		ch = c
	}

	u.appendLog(fmt.Sprintf("Connecting to %s (%s) on channel %d...", dev.Name, dev.Address, ch))
	u.statusLabel.SetText("Status: Connecting...")

	go func() {
		err := u.client.Connect(dev.Address, uint8(ch))
		if err != nil {
			u.statusLabel.SetText("Status: Connection Failed")
			u.appendLog(fmt.Sprintf("Failed to connect: %v", err))
			dialog.ShowError(err, u.window)
			return
		}

		u.statusLabel.SetText(fmt.Sprintf("Status: Connected to %s", dev.Name))
		u.connectBtn.SetText("Disconnect")
		u.connectBtn.Importance = widget.DangerImportance
		u.appendLog("Connected successfully!")
	}()
}

func (u *UIApp) buildUI() fyne.CanvasObject {
	// 1. Connection Header
	u.deviceSelect = widget.NewSelect([]string{}, nil)
	u.deviceSelect.PlaceHolder = "Select Bluetooth Printer..."

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		u.refreshPairedDevices()
	})

	scanBtn := widget.NewButtonWithIcon("Scan", theme.SearchIcon(), func() {
		u.startScan()
	})

	u.channelEntry = widget.NewEntry()
	u.channelEntry.SetText("1")

	u.widthSelect = widget.NewSelect([]string{"58mm (384 dots)", "80mm (576 dots)"}, nil)
	u.widthSelect.SetSelected("58mm (384 dots)")

	u.statusLabel = widget.NewLabel("Status: Disconnected")
	u.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	u.connectBtn = widget.NewButton("Connect", func() {
		u.toggleConnect()
	})
	u.connectBtn.Importance = widget.HighImportance

	headerForm := container.NewVBox(
		widget.NewLabelWithStyle("Printer Connection (RFCOMM)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, nil, container.NewHBox(refreshBtn, scanBtn), u.deviceSelect),
		container.NewHBox(
			widget.NewLabel("Channel:"),
			container.NewGridWrap(fyne.NewSize(50, 36), u.channelEntry),
			widget.NewLabel("Paper:"),
			u.widthSelect,
			u.connectBtn,
			u.statusLabel,
		),
		widget.NewSeparator(),
	)

	// 2. Tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Text", theme.DocumentCreateIcon(), u.buildTextTab()),
		container.NewTabItemWithIcon("Image", theme.FileImageIcon(), u.buildImageTab()),
		container.NewTabItemWithIcon("QR Code", theme.MediaPlayIcon(), u.buildQRTab()),
		container.NewTabItemWithIcon("Raw Hex", theme.SettingsIcon(), u.buildRawTab()),
		container.NewTabItemWithIcon("Diagnostics", theme.HelpIcon(), u.buildDiagTab()),
	)

	// 3. Activity Log
	u.logText = widget.NewMultiLineEntry()
	u.logText.Disable()
	u.logText.TextStyle = fyne.TextStyle{Monospace: true}
	u.logText.SetMinRowsVisible(6)
	u.logScroll = container.NewVScroll(u.logText)

	logBox := container.NewBorder(
		widget.NewLabelWithStyle("Activity Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		u.logScroll,
	)

	split := container.NewVSplit(tabs, logBox)
	split.SetOffset(0.70)

	return container.NewBorder(headerForm, nil, nil, nil, split)
}

func (u *UIApp) buildTextTab() fyne.CanvasObject {
	u.textEntry = widget.NewMultiLineEntry()
	u.textEntry.SetText("SOREAJA THERMAL PRINTER\nTesting Bluetooth RFCOMM\nFresh Brewed Coffee    $3.50\nCroissant              $2.75\n--------------------------------\nTotal:                 $6.25\n")
	u.textEntry.SetMinRowsVisible(8)

	u.alignSelect = widget.NewRadioGroup([]string{"Left", "Center", "Right"}, nil)
	u.alignSelect.SetSelected("Left")
	u.alignSelect.Horizontal = true

	u.boldCheck = widget.NewCheck("Bold", nil)
	u.underlineCheck = widget.NewCheck("Underline", nil)

	u.fontSizeSelect = widget.NewSelect([]string{"Standard (1x)", "Double Width", "Double Height", "Double Both (2x2)"}, nil)
	u.fontSizeSelect.SetSelected("Standard (1x)")

	u.feedCountEntry = widget.NewEntry()
	u.feedCountEntry.SetText("3")

	printBtn := widget.NewButtonWithIcon("Print Text", theme.DocumentPrintIcon(), func() {
		u.printTextContent(false)
	})
	printBtn.Importance = widget.HighImportance

	printCutBtn := widget.NewButtonWithIcon("Print & Cut", theme.ContentCutIcon(), func() {
		u.printTextContent(true)
	})
	printCutBtn.Importance = widget.SuccessImportance

	cutOnlyBtn := widget.NewButton("Cut Paper", func() {
		if !u.checkConnected() {
			return
		}
		u.appendLog("Cutting paper...")
		_ = u.client.CutPaper(3)
	})

	controls := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Align:"), u.alignSelect,
			u.boldCheck, u.underlineCheck,
			widget.NewLabel("Font:"), u.fontSizeSelect,
		),
		container.NewHBox(
			widget.NewLabel("Feed Lines:"),
			container.NewGridWrap(fyne.NewSize(50, 36), u.feedCountEntry),
			printBtn,
			printCutBtn,
			cutOnlyBtn,
		),
	)

	return container.NewBorder(nil, controls, nil, nil, u.textEntry)
}

func (u *UIApp) printTextContent(cut bool) {
	if !u.checkConnected() {
		return
	}

	u.appendLog("Formatting and printing text...")
	_ = u.client.Init()

	// Align
	switch u.alignSelect.Selected {
	case "Center":
		_ = u.client.SetAlign(printer.AlignCenter)
	case "Right":
		_ = u.client.SetAlign(printer.AlignRight)
	default:
		_ = u.client.SetAlign(printer.AlignLeft)
	}

	// Emphasis
	_ = u.client.SetBold(u.boldCheck.Checked)
	_ = u.client.SetUnderline(u.underlineCheck.Checked)

	// Font Size
	switch u.fontSizeSelect.Selected {
	case "Double Width":
		_ = u.client.SetFontSize(1, 0)
	case "Double Height":
		_ = u.client.SetFontSize(0, 1)
	case "Double Both (2x2)":
		_ = u.client.SetFontSize(1, 1)
	default:
		_ = u.client.SetFontSize(0, 0)
	}

	text := u.textEntry.Text
	_ = u.client.PrintText(text)

	lines := uint8(3)
	if n, err := strconv.Atoi(u.feedCountEntry.Text); err == nil && n >= 0 {
		lines = uint8(n)
	}

	if cut {
		_ = u.client.CutPaper(lines)
		u.appendLog("Text printed and paper cut successfully.")
	} else {
		_ = u.client.Feed(lines)
		u.appendLog("Text printed successfully.")
	}
}

func (u *UIApp) buildImageTab() fyne.CanvasObject {
	u.imagePathEntry = widget.NewEntry()
	u.imagePathEntry.SetPlaceHolder("Path to image file (PNG, JPG)...")

	// Default to test.png if present
	cwd, _ := os.Getwd()
	defaultPng := filepath.Join(cwd, "test.png")
	if _, err := os.Stat(defaultPng); err == nil {
		u.imagePathEntry.SetText(defaultPng)
	}

	u.imagePreview = canvas.NewImageFromFile(u.imagePathEntry.Text)
	u.imagePreview.FillMode = canvas.ImageFillContain
	u.imagePreview.SetMinSize(fyne.NewSize(200, 200))

	browseBtn := widget.NewButtonWithIcon("Browse...", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()
			path := reader.URI().Path()
			u.imagePathEntry.SetText(path)
			u.imagePreview.File = path
			u.imagePreview.Refresh()
		}, u.window)
	})

	u.ditherCheck = widget.NewCheck("Floyd-Steinberg Dithering (Best for photos & gradients)", nil)
	u.ditherCheck.SetChecked(true)

	printImgBtn := widget.NewButtonWithIcon("Print Image", theme.DocumentPrintIcon(), func() {
		if !u.checkConnected() {
			return
		}
		path := u.imagePathEntry.Text
		if path == "" {
			dialog.ShowInformation("No Image", "Please select an image file first.", u.window)
			return
		}
		width := u.getPaperWidth()
		dither := u.ditherCheck.Checked

		u.appendLog(fmt.Sprintf("Rasterizing and printing '%s' (max %d dots, dither=%v)...", filepath.Base(path), width, dither))
		go func() {
			_ = u.client.Init()
			if err := u.client.PrintImage(path, width, dither); err != nil {
				u.appendLog(fmt.Sprintf("Image print error: %v", err))
				dialog.ShowError(err, u.window)
				return
			}
			_ = u.client.CutPaper(3)
			u.appendLog("Image printed successfully!")
		}()
	})
	printImgBtn.Importance = widget.HighImportance

	top := container.NewVBox(
		container.NewBorder(nil, nil, nil, browseBtn, u.imagePathEntry),
		u.ditherCheck,
		printImgBtn,
	)

	return container.NewBorder(top, nil, nil, nil, u.imagePreview)
}

func (u *UIApp) buildQRTab() fyne.CanvasObject {
	u.qrTextEntry = widget.NewEntry()
	u.qrTextEntry.SetText("https://github.com/punkestu/soreaja-thermal")

	u.qrSizeLabel = widget.NewLabel("Module Size: 6")
	u.qrSizeSlider = widget.NewSlider(1, 16)
	u.qrSizeSlider.SetValue(6)
	u.qrSizeSlider.OnChanged = func(v float64) {
		u.qrSizeLabel.SetText(fmt.Sprintf("Module Size: %d", int(v)))
	}

	printQRBtn := widget.NewButtonWithIcon("Print QR Code", theme.DocumentPrintIcon(), func() {
		if !u.checkConnected() {
			return
		}
		text := u.qrTextEntry.Text
		if text == "" {
			dialog.ShowInformation("Empty QR", "Please enter text or URL for QR code.", u.window)
			return
		}
		size := uint8(u.qrSizeSlider.Value)

		u.appendLog(fmt.Sprintf("Printing QR Code (size %d)...", size))
		_ = u.client.Init()
		_ = u.client.SetAlign(printer.AlignCenter)
		if err := u.client.PrintQR(text, size); err != nil {
			u.appendLog(fmt.Sprintf("QR error: %v", err))
			dialog.ShowError(err, u.window)
			return
		}
		_ = u.client.CutPaper(3)
		u.appendLog("QR Code printed successfully.")
	})
	printQRBtn.Importance = widget.HighImportance

	return container.NewVBox(
		widget.NewLabel("QR Content (Text or URL):"),
		u.qrTextEntry,
		container.NewHBox(u.qrSizeLabel, container.NewGridWrap(fyne.NewSize(300, 36), u.qrSizeSlider)),
		printQRBtn,
	)
}

func (u *UIApp) buildRawTab() fyne.CanvasObject {
	u.rawHexEntry = widget.NewMultiLineEntry()
	u.rawHexEntry.SetText("1B 40 1B 61 01 48 65 6C 6C 6F 20 57 6F 72 6C 64 0A 1D 56 42 00")
	u.rawHexEntry.SetMinRowsVisible(4)

	presetsSelect := widget.NewSelect([]string{
		"Initialize Printer (1B 40)",
		"Cut Paper (1D 56 42 00)",
		"Open Cash Drawer (1B 70 00 19 FA)",
		"Sample 'Hello World' (Centered + Cut)",
	}, func(selected string) {
		switch selected {
		case "Initialize Printer (1B 40)":
			u.rawHexEntry.SetText("1B 40")
		case "Cut Paper (1D 56 42 00)":
			u.rawHexEntry.SetText("1B 64 03 1D 56 42 00")
		case "Open Cash Drawer (1B 70 00 19 FA)":
			u.rawHexEntry.SetText("1B 70 00 19 FA")
		case "Sample 'Hello World' (Centered + Cut)":
			u.rawHexEntry.SetText("1B 40 1B 61 01 48 65 6C 6C 6F 20 57 6F 72 6C 64 0A 1D 56 42 00")
		}
	})
	presetsSelect.PlaceHolder = "Load Command Preset..."

	sendBtn := widget.NewButtonWithIcon("Send Raw Bytes", theme.MailSendIcon(), func() {
		if !u.checkConnected() {
			return
		}
		raw := strings.ReplaceAll(u.rawHexEntry.Text, " ", "")
		raw = strings.ReplaceAll(raw, "0x", "")
		raw = strings.ReplaceAll(raw, ",", "")
		raw = strings.ReplaceAll(raw, "\n", "")

		bytes, err := hex.DecodeString(raw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid hex string: %w", err), u.window)
			return
		}

		u.appendLog(fmt.Sprintf("Sending %d raw bytes...", len(bytes)))
		if err := u.client.SendRaw(bytes); err != nil {
			u.appendLog(fmt.Sprintf("Raw send error: %v", err))
			dialog.ShowError(err, u.window)
			return
		}
		u.appendLog("Raw bytes sent successfully.")
	})
	sendBtn.Importance = widget.HighImportance

	return container.NewVBox(
		widget.NewLabel("Preset Commands:"),
		presetsSelect,
		widget.NewLabel("Hex Bytes:"),
		u.rawHexEntry,
		sendBtn,
	)
}

func (u *UIApp) buildDiagTab() fyne.CanvasObject {
	diagBtn := widget.NewButtonWithIcon("Print Diagnostic Test Receipt", theme.DocumentPrintIcon(), func() {
		if !u.checkConnected() {
			return
		}
		width := u.getPaperWidth()
		u.appendLog(fmt.Sprintf("Printing full diagnostic receipt (%d dots)...", width))

		go func() {
			if err := u.client.PrintTestReceipt(width); err != nil {
				u.appendLog(fmt.Sprintf("Diagnostic print error: %v", err))
				dialog.ShowError(err, u.window)
				return
			}
			u.appendLog("Diagnostic test receipt printed successfully!")
		}()
	})
	diagBtn.Importance = widget.SuccessImportance

	infoCard := widget.NewCard(
		"Diagnostic Test Receipt",
		"Tests ESC/POS thermal printer functionality",
		widget.NewLabel("Prints:\n• Alignment test (Left, Center, Right)\n• Font scaling (1x, 2x Width, 2x Height)\n• Style attributes (Bold, Underline)\n• Model 2 QR Code verification\n• Automatic paper cut"),
	)

	return container.NewVBox(
		infoCard,
		diagBtn,
	)
}

func (u *UIApp) checkConnected() bool {
	if !u.client.IsConnected() {
		dialog.ShowInformation("Not Connected", "Please connect to a thermal printer first.", u.window)
		return false
	}
	return true
}

// Ensure escpos is imported
var _ = escpos.AlignLeft
