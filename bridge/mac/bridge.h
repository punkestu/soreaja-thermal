#ifndef PRINTER_BRIDGE_MAC_H
#define PRINTER_BRIDGE_MAC_H

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Device discovery & inquiry
// Invokes onFound(name, macAddress) for each discovered/paired device
void printer_scan_devices(void (*onFound)(const char *name, const char *address));
void printer_get_paired_devices(void (*onFound)(const char *name, const char *address));

// Connection management
bool printer_connect(const char *macAddress);
bool printer_connect_channel(const char *macAddress, uint8_t channelID);
void printer_disconnect(void);
bool printer_is_connected(void);

// Raw byte transmission over RFCOMM
bool printer_send_raw(const uint8_t *bytes, int32_t length);

#ifdef __cplusplus
}
#endif

#endif // PRINTER_BRIDGE_MAC_H
