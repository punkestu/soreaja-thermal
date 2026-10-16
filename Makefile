DYLIB_NAME = bridge/mac/libprinterbridge.dylib
CLI_NAME = bin/printer-cli
UI_NAME = bin/printer-ui
APP_NAME = PrinterBridge.app

.PHONY: all bridge cli ui app test clean run-ui run-cli help

all: bridge cli ui app

bridge:
	@echo "==> Building macOS Bluetooth bridge dylib..."
	@$(MAKE) -C bridge/mac

cli: bridge
	@echo "==> Building Go CLI binary ($(CLI_NAME))..."
	@mkdir -p bin
	CGO_ENABLED=1 go build -o $(CLI_NAME) ./cmd/cli
	@echo "==> Built $(CLI_NAME)"

ui: bridge
	@echo "==> Building Go UI binary ($(UI_NAME))..."
	@mkdir -p bin
	CGO_ENABLED=1 go build -o $(UI_NAME) ./cmd/ui
	@echo "==> Built $(UI_NAME)"

app: ui
	@echo "==> Packaging macOS App bundle ($(APP_NAME))..."
	@mkdir -p $(APP_NAME)/Contents/MacOS
	@mkdir -p $(APP_NAME)/Contents/Frameworks
	@mkdir -p $(APP_NAME)/Contents/Resources
	@cp $(UI_NAME) $(APP_NAME)/Contents/MacOS/$(APP_NAME)
	@cp $(DYLIB_NAME) $(APP_NAME)/Contents/Frameworks/
	@cp Info.plist $(APP_NAME)/Contents/Info.plist
	@echo "==> $(APP_NAME) created successfully."

test: bridge
	@echo "==> Running Go unit and integration tests..."
	CGO_ENABLED=1 go test -v ./...

run-ui: app
	@open $(APP_NAME)

run-cli: cli
	@./$(CLI_NAME) help

clean:
	@echo "==> Cleaning build artifacts..."
	@$(MAKE) -C bridge/mac clean
	@rm -rf bin/ $(APP_NAME) go-printer-test printerbridge libprinterbridge.dylib PrinterBridge.h
	@echo "==> Clean complete."
