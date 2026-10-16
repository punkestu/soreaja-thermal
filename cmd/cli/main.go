package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/punkestu/soreaja-thermal/pkg/printer"
)

func printHelp() {
	fmt.Println(`=======================================================
   Soreaja Thermal Printer - Go CLI (ESC/POS RFCOMM)
=======================================================
Usage:
  printer-cli <command> [arguments...]

Commands:
  paired                         List all currently paired Bluetooth devices
  scan                           Scan for nearby Bluetooth devices (5s inquiry)
  test <mac> [width]             Print diagnostic receipt (width: 384 for 58mm, 576 for 80mm)
  text <mac> "<text>" [width]    Print custom text and cut paper
  image <mac> <file> [width]     Print image file with Floyd-Steinberg dithering
  raw <mac> <hex-string>         Send raw hex ESC/POS bytes (e.g. '1B40 1B6101 48690A')
  help                           Show this help message

Examples:
  printer-cli paired
  printer-cli scan
  printer-cli test 86-67-7a-7c-0a-56
  printer-cli text 86-67-7a-7c-0a-56 "Hello from Go!\nLine 2\n"
  printer-cli image 86-67-7a-7c-0a-56 test.png 384
  printer-cli raw 86-67-7a-7c-0a-56 "1B40 1B4501 48656C6C6F0A 1D564200"`)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	command := strings.ToLower(args[0])

	switch command {
	case "help", "-h", "--help":
		printHelp()

	case "paired", "list":
		listPaired()

	case "scan":
		scanDevices()

	case "test":
		if len(args) < 2 {
			fmt.Println("Error: Missing MAC address. Usage: printer-cli test <mac> [width]")
			os.Exit(1)
		}
		mac := args[1]
		width := 384
		if len(args) >= 3 {
			if w, err := strconv.Atoi(args[2]); err == nil && w > 0 {
				width = w
			}
		}
		runTestReceipt(mac, width)

	case "text":
		if len(args) < 3 {
			fmt.Println("Error: Missing arguments. Usage: printer-cli text <mac> \"<text>\" [width]")
			os.Exit(1)
		}
		mac := args[1]
		text := args[2]
		runPrintText(mac, text)

	case "image":
		if len(args) < 3 {
			fmt.Println("Error: Missing arguments. Usage: printer-cli image <mac> <file_path> [width]")
			os.Exit(1)
		}
		mac := args[1]
		path := args[2]
		width := 384
		if len(args) >= 4 {
			if w, err := strconv.Atoi(args[3]); err == nil && w > 0 {
				width = w
			}
		}
		runPrintImage(mac, path, width)

	case "raw":
		if len(args) < 3 {
			fmt.Println("Error: Missing arguments. Usage: printer-cli raw <mac> <hex-string>")
			os.Exit(1)
		}
		mac := args[1]
		hexStr := args[2]
		runSendRaw(mac, hexStr)

	default:
		fmt.Printf("Unknown command: %q. Run 'printer-cli help' for options.\n", command)
		os.Exit(1)
	}
}

func listPaired() {
	fmt.Println("\nSearching paired Bluetooth devices...")
	devices, err := printer.GetPairedDevices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if len(devices) == 0 {
		fmt.Println("No paired Bluetooth devices found.")
		fmt.Println("Tip: Pair your thermal printer in macOS System Settings -> Bluetooth first.")
		return
	}
	fmt.Printf("Found %d paired device(s):\n", len(devices))
	for i, dev := range devices {
		fmt.Printf(" [%d] %-32s | %s\n", i+1, dev.Name, dev.Address)
	}
}

func scanDevices() {
	fmt.Println("\nScanning for Bluetooth devices for 5 seconds...")
	count := 0
	devices, err := printer.ScanDevices(func(dev printer.DeviceInfo) {
		count++
		fmt.Printf(" [%d] Discovered: %-30s | %s\n", count, dev.Name, dev.Address)
	})
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		return
	}
	fmt.Printf("\nScan complete. Discovered %d total device(s).\n", len(devices))
}

func runTestReceipt(mac string, width int) {
	fmt.Printf("\nConnecting to printer at %s...\n", mac)
	client, err := printer.Connect(mac, 1)
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	fmt.Println("Connected! Printing diagnostic receipt...")
	if err := client.PrintTestReceipt(width); err != nil {
		fmt.Printf("Print error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Test receipt printed successfully! Disconnecting.")
}

func runPrintText(mac, text string) {
	fmt.Printf("\nConnecting to printer at %s...\n", mac)
	client, err := printer.Connect(mac, 1)
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	fmt.Println("Connected! Sending text to printer...")
	_ = client.Init()
	// Process escaped newlines (\n) if passed from shell
	processedText := strings.ReplaceAll(text, `\n`, "\n")
	processedText = strings.ReplaceAll(processedText, `\t`, "\t")
	_ = client.PrintText(processedText)
	_ = client.CutPaper(3)
	fmt.Println("Text printed and cut successfully! Disconnecting.")
}

func runPrintImage(mac, path string, width int) {
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("Error: image file not found at '%s'\n", path)
		os.Exit(1)
	}

	fmt.Printf("\nConnecting to printer at %s...\n", mac)
	client, err := printer.Connect(mac, 1)
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	fmt.Printf("Connected! Rasterizing and printing '%s' (%d dots wide)...\n", path, width)
	_ = client.Init()
	if err := client.PrintImage(path, width, true); err != nil {
		fmt.Printf("Image print error: %v\n", err)
		os.Exit(1)
	}
	_ = client.CutPaper(3)
	fmt.Println("Image printed successfully! Disconnecting.")
}

func runSendRaw(mac, hexStr string) {
	cleanHex := strings.ReplaceAll(hexStr, " ", "")
	cleanHex = strings.ReplaceAll(cleanHex, "0x", "")
	cleanHex = strings.ReplaceAll(cleanHex, ",", "")

	data, err := hex.DecodeString(cleanHex)
	if err != nil {
		fmt.Printf("Error: Invalid hex string '%s': %v\n", hexStr, err)
		os.Exit(1)
	}

	fmt.Printf("\nConnecting to printer at %s...\n", mac)
	client, err := printer.Connect(mac, 1)
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	fmt.Printf("Connected! Sending %d raw bytes...\n", len(data))
	if err := client.SendRaw(data); err != nil {
		fmt.Printf("Send error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Raw bytes sent successfully! Disconnecting.")
}
