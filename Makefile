APP_NAME := pvvc
BUILD_DIR := dist
ROOT_PKG := github.com/4okimi7uki/pvvc/cmd
VERSION ?= v0.0.0-dev

.PHONY: default build build-all clean

default: build

build:
	@echo "🚀 Building for your current OS (version: $(VERSION))"
	go build -ldflags "-X $(ROOT_PKG).version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) .

build-all:
	@echo "📦 Building all platforms (version: $(VERSION))"

	GOOS=darwin GOARCH=arm64 go build -ldflags "-X $(ROOT_PKG).version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)_mac_arm64 .

	GOOS=darwin GOARCH=amd64 go build -ldflags "-X $(ROOT_PKG).version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)_mac_amd64 .

	GOOS=linux GOARCH=amd64 go build -ldflags "-X $(ROOT_PKG).version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)_linux_amd64 .

	GOOS=windows GOARCH=amd64 go build -ldflags "-X $(ROOT_PKG).version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME).exe .

	@echo "--------------------------------------"
	@echo " 🎉 All binaries built successfully!"
	@echo "--------------------------------------"

clean:
	@rm -rf $(BUILD_DIR)
	@echo "🧹 Cleaned build directory"
