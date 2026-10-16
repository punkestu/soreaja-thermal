SWIFTC ?= swiftc
FRAMEWORKS = -framework IOBluetooth -framework Foundation
DYLIB_NAME = libprinterbridge.dylib

.PHONY: all clean

all: $(DYLIB_NAME)

$(DYLIB_NAME): PrinterBridge.swift
	@echo "==> Building macOS Bluetooth bridge $(DYLIB_NAME)..."
	@mkdir -p .module-cache
	$(SWIFTC) $(FRAMEWORKS) -module-cache-path .module-cache -Xlinker -install_name -Xlinker @rpath/$(DYLIB_NAME) -emit-library PrinterBridge.swift -o $(DYLIB_NAME)
	@echo "==> Built $(DYLIB_NAME) successfully."

clean:
	rm -rf $(DYLIB_NAME) .module-cache
