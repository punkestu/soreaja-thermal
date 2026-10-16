package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/punkestu/soreaja-thermal/pkg/printer"
)

func main() {
	macFlag := flag.String("mac", "", "Printer Bluetooth MAC address (e.g. 86-67-7a-7c-0a-56)")
	channelFlag := flag.Int("channel", 1, "RFCOMM channel ID (default: 1)")
	widthFlag := flag.Int("width", 384, "Paper width in dots (384 for 58mm, 576 for 80mm)")
	textFlag := flag.String("text", "", "Custom text to print instead of diagnostic receipt")
	imageFlag := flag.String("image", "", "Path to image to print (e.g. test.png)")
	listFlag := flag.Bool("list", false, "List paired Bluetooth devices and exit")
	scanFlag := flag.Bool("scan", false, "Scan for nearby Bluetooth devices (5s inquiry) and exit")
	flag.Parse()

	if *listFlag {
		listPaired()
		return
	}

	if *scanFlag {
		scanDevices()
		return
	}

	targetMAC := *macFlag

	// If no MAC address specified, automatically search for paired devices
	if targetMAC == "" {
		fmt.Println("No -mac flag specified. Checking paired devices for thermal printer...")
		devices, err := printer.GetPairedDevices()
		if err != nil {
			fmt.Printf("Error fetching paired devices: %v\n", err)
			os.Exit(1)
		}

		for _, dev := range devices {
			if strings.EqualFold(dev.Name, "RPP02N") || strings.Contains(strings.ToLower(dev.Name), "printer") {
				targetMAC = dev.Address
				fmt.Printf("=> Auto-detected printer: %s (%s)\n", dev.Name, dev.Address)
				break
			}
		}

		if targetMAC == "" {
			if len(devices) > 0 {
				fmt.Println("\nAvailable paired devices:")
				for i, dev := range devices {
					fmt.Printf(" [%d] %-30s | %s\n", i+1, dev.Name, dev.Address)
				}
				fmt.Println("\nPlease run with -mac <address> or pair your printer in macOS Bluetooth settings.")
			} else {
				fmt.Println("No paired Bluetooth devices found. Please pair your printer in macOS Settings.")
			}
			os.Exit(1)
		}
	}

	fmt.Printf("\nConnecting to printer at %s (RFCOMM channel %d)...\n", targetMAC, *channelFlag)
	p, err := printer.Connect(targetMAC, uint8(*channelFlag))
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer p.Disconnect()

	fmt.Println("Connected successfully!")

	// 1. If custom text requested
	if *textFlag != "" {
		fmt.Printf("Sending text to printer...\n")
		_ = p.Init()
		_ = p.PrintLine(*textFlag)
		_ = p.CutPaper(3)
		fmt.Println("Text printed and paper cut!")
		return
	}

	// 2. If specific image requested
	if *imageFlag != "" {
		fmt.Printf("Printing image '%s' (%d dots wide)...\n", *imageFlag, *widthFlag)
		_ = p.Init()
		err := p.PrintImage(*imageFlag, *widthFlag, true)
		if err != nil {
			fmt.Printf("Error printing image: %v\n", err)
		} else {
			_ = p.CutPaper(3)
			fmt.Println("Image printed successfully!")
		}
		return
	}

	// 3. Default: Print full diagnostic test receipt
	fmt.Printf("Printing diagnostic test receipt (width: %d dots)...\n", *widthFlag)
	if err := p.PrintTestReceipt(*widthFlag); err != nil {
		fmt.Printf("Print error: %v\n", err)
		os.Exit(1)
	}

	// 4. Also test image printing if test.png exists in current directory
	if _, err := os.Stat("test.png"); err == nil {
		fmt.Println("Found test.png, printing sample receipt image...")
		_ = p.Init()
		if err := p.PrintImage("test.png", *widthFlag, true); err != nil {
			fmt.Printf("Warning: test.png print failed: %v\n", err)
		} else {
			_ = p.CutPaper(3)
			fmt.Println("test.png printed successfully!")
		}
	}

	fmt.Println("All print jobs completed. Disconnecting.")
}

func listPaired() {
	fmt.Println("\nPaired Bluetooth Devices:")
	fmt.Println(strings.Repeat("-", 60))
	devices, err := printer.GetPairedDevices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for i, dev := range devices {
		fmt.Printf(" [%d] %-30s | %s\n", i+1, dev.Name, dev.Address)
	}
	fmt.Println(strings.Repeat("-", 60))
}

func scanDevices() {
	fmt.Println("\nScanning for Bluetooth devices for 5 seconds...")
	count := 0
	devices, err := printer.ScanDevices(func(dev printer.DeviceInfo) {
		count++
		fmt.Printf(" [%d] Found: %s (%s)\n", count, dev.Name, dev.Address)
	})
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		return
	}
	fmt.Printf("Scan finished. Discovered %d device(s).\n", len(devices))
}
