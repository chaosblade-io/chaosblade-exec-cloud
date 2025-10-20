# Copyright 2025 The ChaosBlade Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: build clean linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64 help build_all

BLADE_SRC_ROOT=$(shell pwd)
UNAME := $(shell uname)

# Version management - auto-detect from git tags
ifeq ($(BLADE_VERSION), )
	BLADE_VERSION=$(shell ./version/version.sh version)
endif

# Additional version information
GIT_COMMIT=$(shell ./version/version.sh commit)
BUILD_TIME=$(shell ./version/version.sh build-time)
BUILD_TYPE=$(shell ./version/version.sh build-type)
FULL_VERSION=$(shell ./version/version.sh full-version)

BUILD_TARGET=target
BUILD_TARGET_DIR_NAME=chaosblade-$(BLADE_VERSION)
BUILD_TARGET_PKG_DIR=$(BUILD_TARGET)/chaosblade-$(BLADE_VERSION)
BUILD_TARGET_BIN=$(BUILD_TARGET_PKG_DIR)/bin
BUILD_TARGET_YAML=$(BUILD_TARGET_PKG_DIR)/yaml

# Platform-specific directory functions
define get_platform_dir_name
chaosblade-$(BLADE_VERSION)-$(1)
endef

define get_platform_pkg_dir
$(BUILD_TARGET)/chaosblade-$(BLADE_VERSION)-$(1)
endef

define get_platform_bin_dir
$(BUILD_TARGET)/chaosblade-$(BLADE_VERSION)-$(1)/bin
endef

define get_platform_yaml_dir
$(BUILD_TARGET)/chaosblade-$(BLADE_VERSION)-$(1)/yaml
endef

# YAML file name
CLOUD_YAML_FILE_NAME=chaosblade-cloud-spec-$(BLADE_VERSION).yaml
CLOUD_YAML_FILE_PATH=$(BUILD_TARGET_YAML)/$(CLOUD_YAML_FILE_NAME)

# Binary file name
CLOUD_BINARY_NAME=chaos_cloud

GO_ENV=CGO_ENABLED=0
GO_MODULE=GO111MODULE=on
GO=env $(GO_ENV) $(GO_MODULE) go

# Cross-compilation GO command (without CGO)
GO_CROSS=env CGO_ENABLED=0 GO111MODULE=on go

# Host platform for running generators (ensure go run executes on host arch)
HOST_GOOS=$(shell go env GOHOSTOS)
HOST_GOARCH=$(shell go env GOHOSTARCH)

# Build flags for different platforms with version information
GO_FLAGS_LINUX_AMD64=-ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE) -s -w"

GO_FLAGS_LINUX_ARM64=-ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE) -s -w"

GO_FLAGS_DARWIN_AMD64=-ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE) -s -w"

GO_FLAGS_DARWIN_ARM64=-ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE) -s -w"


# Common build flags
GO_FLAGS_COMMON=-ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE) -s -w"

# Help target (default)
help:
	@echo "ChaosBlade Cloud Executor Build System"
	@echo "======================================"
	@echo ""
	@echo "Version Information:"
	@echo "  Version:         $(BLADE_VERSION)"
	@echo "  Git Commit:      $(GIT_COMMIT)"
	@echo "  Build Time:      $(BUILD_TIME)"
	@echo "  Build Type:      $(BUILD_TYPE)"
	@echo "  Full Version:    $(FULL_VERSION)"
	@echo ""
	@echo "Available Build Targets:"
	@echo "  build          - Build current platform version"
	@echo "  build_all      - Build all platform versions"
	@echo "  linux_amd64    - Build Linux AMD64 version"
	@echo "  linux_arm64    - Build Linux ARM64 version"
	@echo "  darwin_amd64   - Build macOS AMD64 version"
	@echo "  darwin_arm64   - Build macOS ARM64 version"
	@echo ""
	@echo "Other Commands:"
	@echo "  test           - Run tests"
	@echo "  format         - Format Go code using goimports and gofumpt"
	@echo "  verify         - Verify Go code formatting and import order"
	@echo "  license-check  - Check license headers"
	@echo "  clean          - Clean build products"
	@echo "  all            - Build and test"
	@echo "  help           - Show this help information"
	@echo "  version        - Show version information"
	@echo ""
	@echo "Environment Variables:"
	@echo "  BLADE_VERSION  - Specify build version (default: auto-detect from Git Tag)"
	@echo ""
	@echo "Usage Examples:"
	@echo "  make help                    # Show help"
	@echo "  make build                   # Build current platform version"
	@echo "  make build_all               # Build all platform versions"
	@echo "  make linux_amd64            # Build Linux AMD64 version"
	@echo "  BLADE_VERSION=1.8.0 make build  # Build with specified version"
	@echo ""

# Default target
.DEFAULT_GOAL := help

# Version info target
version:
	@echo "Version Information:"
	@echo "  Version:         $(BLADE_VERSION)"
	@echo "  Git Commit:      $(GIT_COMMIT)"
	@echo "  Build Time:      $(BUILD_TIME)"
	@echo "  Build Type:      $(BUILD_TYPE)"
	@echo "  Full Version:    $(FULL_VERSION)"
	@echo "  Is Tagged:       $(shell ./version/version.sh is-tagged)"

# build cloud for current platform
build: pre_build build_yaml build_cloud

pre_build:
	rm -rf $(BUILD_TARGET_PKG_DIR)
	mkdir -p $(BUILD_TARGET_BIN) $(BUILD_TARGET_YAML)

build_yaml: build/spec.go
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) $(GO) run -ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE)" $< $(CLOUD_YAML_FILE_PATH)

build_cloud: main.go
	$(GO) build $(GO_FLAGS_COMMON) -o $(BUILD_TARGET_BIN)/$(CLOUD_BINARY_NAME) $<

# Multi-platform build targets
linux_amd64:
	$(eval PLATFORM := linux_amd64)
	$(eval PLATFORM_PKG_DIR := $(call get_platform_pkg_dir,$(PLATFORM)))
	$(eval PLATFORM_BIN_DIR := $(call get_platform_bin_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_DIR := $(call get_platform_yaml_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_FILE := $(PLATFORM_YAML_DIR)/$(CLOUD_YAML_FILE_NAME))
	rm -rf $(PLATFORM_PKG_DIR)
	mkdir -p $(PLATFORM_BIN_DIR) $(PLATFORM_YAML_DIR)
	GOOS=linux GOARCH=amd64 $(GO_CROSS) build $(GO_FLAGS_LINUX_AMD64) -o $(PLATFORM_BIN_DIR)/$(CLOUD_BINARY_NAME) main.go
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) $(GO) run -ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE)" build/spec.go $(PLATFORM_YAML_FILE)

linux_arm64:
	$(eval PLATFORM := linux_arm64)
	$(eval PLATFORM_PKG_DIR := $(call get_platform_pkg_dir,$(PLATFORM)))
	$(eval PLATFORM_BIN_DIR := $(call get_platform_bin_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_DIR := $(call get_platform_yaml_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_FILE := $(PLATFORM_YAML_DIR)/$(CLOUD_YAML_FILE_NAME))
	rm -rf $(PLATFORM_PKG_DIR)
	mkdir -p $(PLATFORM_BIN_DIR) $(PLATFORM_YAML_DIR)
	GOOS=linux GOARCH=arm64 $(GO_CROSS) build $(GO_FLAGS_LINUX_ARM64) -o $(PLATFORM_BIN_DIR)/$(CLOUD_BINARY_NAME) main.go
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) $(GO) run -ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE)" build/spec.go $(PLATFORM_YAML_FILE)

darwin_amd64:
	$(eval PLATFORM := darwin_amd64)
	$(eval PLATFORM_PKG_DIR := $(call get_platform_pkg_dir,$(PLATFORM)))
	$(eval PLATFORM_BIN_DIR := $(call get_platform_bin_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_DIR := $(call get_platform_yaml_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_FILE := $(PLATFORM_YAML_DIR)/$(CLOUD_YAML_FILE_NAME))
	rm -rf $(PLATFORM_PKG_DIR)
	mkdir -p $(PLATFORM_BIN_DIR) $(PLATFORM_YAML_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(GO_FLAGS_DARWIN_AMD64) -o $(PLATFORM_BIN_DIR)/$(CLOUD_BINARY_NAME) main.go
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) $(GO) run -ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE)" build/spec.go $(PLATFORM_YAML_FILE)

darwin_arm64:
	$(eval PLATFORM := darwin_arm64)
	$(eval PLATFORM_PKG_DIR := $(call get_platform_pkg_dir,$(PLATFORM)))
	$(eval PLATFORM_BIN_DIR := $(call get_platform_bin_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_DIR := $(call get_platform_yaml_dir,$(PLATFORM)))
	$(eval PLATFORM_YAML_FILE := $(PLATFORM_YAML_DIR)/$(CLOUD_YAML_FILE_NAME))
	rm -rf $(PLATFORM_PKG_DIR)
	mkdir -p $(PLATFORM_BIN_DIR) $(PLATFORM_YAML_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GO_FLAGS_DARWIN_ARM64) -o $(PLATFORM_BIN_DIR)/$(CLOUD_BINARY_NAME) main.go
	GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH) $(GO) run -ldflags="-X main.Version=$(BLADE_VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.BuildType=$(BUILD_TYPE)" build/spec.go $(PLATFORM_YAML_FILE)


# test
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# clean all build result
clean:
	go clean ./...
	rm -rf $(BUILD_TARGET)

# Build all platforms
build_all: linux_amd64 linux_arm64 darwin_amd64 darwin_arm64
	@echo "=========================================="
	@echo "All platform builds completed successfully!"
	@echo "Generated directories:"
	@echo "  - chaosblade-$(BLADE_VERSION)-linux_amd64"
	@echo "  - chaosblade-$(BLADE_VERSION)-linux_arm64"
	@echo "  - chaosblade-$(BLADE_VERSION)-darwin_amd64"
	@echo "  - chaosblade-$(BLADE_VERSION)-darwin_arm64"
	@echo "=========================================="

all: build test

.PHONY: format
format: license-format
	@echo "Running goimports and gofumpt to format Go code..."
	@./hack/update-imports.sh
	@./hack/update-gofmt.sh

.PHONY: verify
verify:
	@echo "Verifying Go code formatting and import order..."
	@./hack/verify-gofmt.sh
	@./hack/verify-imports.sh

.PHONY: license-check
license-check:
	@echo "Checking license headers..."
	docker run -it --rm -v $(shell pwd):/github/workspace ghcr.io/korandoru/hawkeye check

.PHONY: license-format
license-format:
	@echo "Formatting license headers..."
	docker run -it --rm -v $(shell pwd):/github/workspace ghcr.io/korandoru/hawkeye format
