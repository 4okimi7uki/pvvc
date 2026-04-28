APP_NAME := pvvc
BUILD_DIR := dist
ROOT_PKG := github.com/4okimi7uki/pvvc
VER_PKG := github.com/4okimi7uki/pvvc/internal/gh
VERSION ?= v0.0.0-dev

-include Makefile.local

.PHONY: default build build-all clean sync-org

default: build

build:
	@echo "🚀 Building for your current OS (version: $(VERSION))"
	go build -ldflags "-X $(VER_PKG).BuildVersion=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) .

build-all:
	@echo "📦 Building all platforms (version: $(VERSION))"
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X $(VER_PKG).BuildVersion=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)_mac_arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X $(VER_PKG).BuildVersion=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)_mac_amd64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "-X $(VER_PKG).BuildVersion=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME)_linux_amd64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-X $(VER_PKG).BuildVersion=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME).exe .
	@echo "--------------------------------------"
	@echo " 🎉 All binaries built successfully!"
	@echo "--------------------------------------"

clean:
	@rm -rf $(BUILD_DIR)
	@echo "🧹 Cleaned build directory"

sync-org:
ifndef ORG_PKG
	$(error ORG_PKG is not set. Create a Makefile.local with ORG_PKG and ORG_DIR)
endif
	@echo "🔄 Syncing changes to org repo"
	rsync -av --exclude='.git' --exclude='Makefile' --exclude='go.mod' --exclude='go.sum' . $(ORG_DIR)/
	find $(ORG_DIR) -name "*.go" | xargs sed -i '' "s|$(ROOT_PKG)|$(ORG_PKG)|g"
	cd $(ORG_DIR) && go mod edit -module $(ORG_PKG)
	@echo "✅ Sync done"
