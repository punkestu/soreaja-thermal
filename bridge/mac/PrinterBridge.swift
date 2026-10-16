import Foundation
import IOBluetooth

// MARK: - Bluetooth Device Model

public struct BluetoothDeviceInfo {
    public let name: String
    public let address: String
    public let isPaired: Bool
}

// MARK: - Bluetooth Scanner Manager

public class ScannerManager: NSObject, IOBluetoothDeviceInquiryDelegate {
    public static let shared = ScannerManager()
    
    private var inquiry: IOBluetoothDeviceInquiry?
    public var onDeviceFound: ((BluetoothDeviceInfo) -> Void)?
    public var onScanFinished: (() -> Void)?
    public private(set) var isScanning = false
    
    public var cCallback: (@convention(c) (UnsafePointer<CChar>?, UnsafePointer<CChar>?) -> Void)?

    public func startScanSyncC(duration: UInt8 = 5, onDeviceFound: @escaping @convention(c) (UnsafePointer<CChar>?, UnsafePointer<CChar>?) -> Void) {
        self.cCallback = onDeviceFound
        self.onDeviceFound = nil
        stopScan()
        
        inquiry = IOBluetoothDeviceInquiry(delegate: self)
        inquiry?.updateNewDeviceNames = true
        inquiry?.inquiryLength = duration
        isScanning = true
        
        if inquiry?.start() == kIOReturnSuccess {
            while isScanning {
                RunLoop.current.run(mode: .default, before: Date(timeIntervalSinceNow: 0.1))
            }
        } else {
            isScanning = false
        }
    }
    
    public func stopScan() {
        if isScanning {
            inquiry?.stop()
            isScanning = false
        }
    }

    public func deviceInquiryDeviceFound(_ sender: IOBluetoothDeviceInquiry!, device: IOBluetoothDevice!) {
        handleDevice(device)
    }
    
    public func deviceInquiryDeviceNameUpdated(_ sender: IOBluetoothDeviceInquiry!, device: IOBluetoothDevice!, devicesRemaining: UInt32) {
        handleDevice(device)
    }

    public func deviceInquiryComplete(_ sender: IOBluetoothDeviceInquiry!, error: IOReturn, aborted: Bool) {
        isScanning = false
        onScanFinished?()
    }

    private func handleDevice(_ device: IOBluetoothDevice!) {
        guard let device = device else { return }
        let name = device.name ?? "Unknown Device"
        let address = device.addressString ?? "Unknown"
        let isPaired = device.isPaired()
        
        let info = BluetoothDeviceInfo(name: name, address: address, isPaired: isPaired)
        onDeviceFound?(info)
        
        if let cb = cCallback {
            name.withCString { cName in
                address.withCString { cAddr in
                    cb(cName, cAddr)
                }
            }
        }
    }
    
    public static func getPairedDevices() -> [BluetoothDeviceInfo] {
        guard let devices = IOBluetoothDevice.pairedDevices() as? [IOBluetoothDevice] else { return [] }
        return devices.map { dev in
            let name = dev.name ?? "Unknown Device"
            let addr = dev.addressString ?? "Unknown"
            return BluetoothDeviceInfo(name: name, address: addr, isPaired: true)
        }
    }
}

// MARK: - Bluetooth RFCOMM Printer Manager

public class PrinterManager: NSObject, IOBluetoothRFCOMMChannelDelegate {
    public static let shared = PrinterManager()
    
    public private(set) var channel: IOBluetoothRFCOMMChannel?
    public private(set) var connectedAddress: String?
    public private(set) var connectedName: String?
    
    public var isConnected: Bool {
        return channel?.isOpen() ?? false
    }
    
    public func connect(macAddress: String, channelID: UInt8 = 1) -> Bool {
        disconnect()
        
        // Normalize address: supports AA:BB:CC:DD:EE:FF, AA-BB-CC-DD-EE-FF, AABBCCDDEEFF
        let cleanedAddr = macAddress.trimmingCharacters(in: .whitespacesAndNewlines)
            .replacingOccurrences(of: ":", with: "-")
        
        guard let device = IOBluetoothDevice(addressString: cleanedAddr) else {
            return false
        }
        
        var newChannel: IOBluetoothRFCOMMChannel?
        let status = device.openRFCOMMChannelSync(&newChannel, withChannelID: BluetoothRFCOMMChannelID(channelID), delegate: self)
        
        if status == kIOReturnSuccess, let ch = newChannel {
            self.channel = ch
            self.connectedAddress = macAddress
            self.connectedName = device.name ?? macAddress
            return true
        } else {
            return false
        }
    }
    
    public func disconnect() {
        if let ch = channel {
            ch.setDelegate(nil)
            _ = ch.close()
            channel = nil
        }
        connectedAddress = nil
        connectedName = nil
    }
    
    /// Writes raw bytes to printer in safe RFCOMM chunks (MTU ~ 512 bytes)
    @discardableResult
    public func send(bytes: [UInt8], chunkSize: Int = 512, delayMs: UInt32 = 15) -> Bool {
        guard let ch = channel, ch.isOpen() else {
            return false
        }
        
        var offset = 0
        while offset < bytes.count {
            let count = min(chunkSize, bytes.count - offset)
            var slice = Array(bytes[offset..<(offset + count)])
            let ret = ch.writeSync(&slice, length: UInt16(slice.count))
            if ret != kIOReturnSuccess {
                return false
            }
            offset += count
            if offset < bytes.count && delayMs > 0 {
                usleep(delayMs * 1000)
            }
        }
        return true
    }

    // IOBluetoothRFCOMMChannelDelegate
    public func rfcommChannelClosed(_ rfcommChannel: IOBluetoothRFCOMMChannel!) {
        disconnect()
    }
}

// MARK: - C ABI Exports

public typealias DeviceFoundCb = @convention(c) (UnsafePointer<CChar>?, UnsafePointer<CChar>?) -> Void

@_cdecl("printer_scan_devices")
public func printer_scan_devices(onFound: @escaping DeviceFoundCb) {
    ScannerManager.shared.startScanSyncC(onDeviceFound: onFound)
}

@_cdecl("printer_get_paired_devices")
public func printer_get_paired_devices(onFound: @escaping DeviceFoundCb) {
    let devices = ScannerManager.getPairedDevices()
    for dev in devices {
        dev.name.withCString { cName in
            dev.address.withCString { cAddr in
                onFound(cName, cAddr)
            }
        }
    }
}

@_cdecl("printer_connect")
public func printer_connect(macAddress: UnsafePointer<CChar>) -> Bool {
    return PrinterManager.shared.connect(macAddress: String(cString: macAddress))
}

@_cdecl("printer_connect_channel")
public func printer_connect_channel(macAddress: UnsafePointer<CChar>, channelID: UInt8) -> Bool {
    return PrinterManager.shared.connect(macAddress: String(cString: macAddress), channelID: channelID)
}

@_cdecl("printer_disconnect")
public func printer_disconnect() {
    PrinterManager.shared.disconnect()
}

@_cdecl("printer_is_connected")
public func printer_is_connected() -> Bool {
    return PrinterManager.shared.isConnected
}

@_cdecl("printer_send_raw")
public func printer_send_raw(bytes: UnsafePointer<UInt8>, length: Int32) -> Bool {
    guard length > 0 else { return true }
    let buffer = Array(UnsafeBufferPointer(start: bytes, count: Int(length)))
    return PrinterManager.shared.send(bytes: buffer)
}
