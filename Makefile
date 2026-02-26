APP_NAME = CryptoBar
BUILD_DIR = build
BINARY = $(BUILD_DIR)/$(APP_NAME)
APP_BUNDLE = $(BUILD_DIR)/$(APP_NAME).app
GO ?= /opt/homebrew/Cellar/go@1.23/1.23.12/bin/go

.PHONY: build run clean install app

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GO) build -o $(BINARY) .

run: build
	$(BINARY)

app: build
	@mkdir -p "$(APP_BUNDLE)/Contents/MacOS"
	@mkdir -p "$(APP_BUNDLE)/Contents/Resources"
	@cp $(BINARY) "$(APP_BUNDLE)/Contents/MacOS/$(APP_NAME)"
	@cp assets/Info.plist "$(APP_BUNDLE)/Contents/Info.plist"
	@cp assets/AppIcon.icns "$(APP_BUNDLE)/Contents/Resources/AppIcon.icns"
	@echo "Built $(APP_BUNDLE)"

install: app
	@cp -r "$(APP_BUNDLE)" /Applications/
	@echo "Installed to /Applications/$(APP_NAME).app"

clean:
	rm -rf $(BUILD_DIR)
