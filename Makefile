APP_NAME = PortSpy
BINARY = $(APP_NAME)
APP_BUNDLE = $(APP_NAME).app
BUNDLE_DIR = $(APP_BUNDLE)/Contents
MACOS_DIR = $(BUNDLE_DIR)/MacOS
RESOURCES_DIR = $(BUNDLE_DIR)/Resources

.PHONY: all build app clean

all: app

build:
	go build -o $(BINARY) .

app: build
	mkdir -p $(MACOS_DIR) $(RESOURCES_DIR)
	cp $(BINARY) $(MACOS_DIR)/$(BINARY)
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $(BUNDLE_DIR)/Info.plist
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(BUNDLE_DIR)/Info.plist
	@echo '<plist version="1.0">' >> $(BUNDLE_DIR)/Info.plist
	@echo '<dict>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>CFBundleExecutable</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <string>$(BINARY)</string>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>CFBundleIdentifier</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <string>com.kyle.portspy</string>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>CFBundleName</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <string>$(APP_NAME)</string>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>CFBundleVersion</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <string>1.0</string>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>CFBundleShortVersionString</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <string>1.0</string>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>LSUIElement</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <true/>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <key>NSHighResolutionCapable</key>' >> $(BUNDLE_DIR)/Info.plist
	@echo '    <true/>' >> $(BUNDLE_DIR)/Info.plist
	@echo '</dict>' >> $(BUNDLE_DIR)/Info.plist
	@echo '</plist>' >> $(BUNDLE_DIR)/Info.plist
	@echo "Built $(APP_BUNDLE)"

clean:
	rm -f $(BINARY)
	rm -rf $(APP_BUNDLE)
