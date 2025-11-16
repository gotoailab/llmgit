.PHONY: build build-all clean install help build-linux build-windows build-darwin build-darwin-arm64

# 构建目录
BUILD_DIR := build
BINARY_NAME := llmgit

# Go 参数
GO := go
GOFLAGS := -v

# 版本信息
VERSION := v0.0.2
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# LDFLAGS 包含版本和构建信息
LDFLAGS := -s -w \
	-X 'github.com/gotoailab/llmgit/internal/commands.Version=$(VERSION)' \
	-X 'github.com/gotoailab/llmgit/internal/commands.BuildDate=$(BUILD_DATE)' \
	-X 'github.com/gotoailab/llmgit/internal/commands.GitCommit=$(GIT_COMMIT)'

# 支持的平台
PLATFORMS := linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64

help: ## 显示帮助信息
	@echo "可用的命令:"
	@echo "  make build           - 构建当前平台的项目"
	@echo "  make build-all       - 构建所有平台的项目"
	@echo "  make build-linux     - 构建 Linux 平台 (amd64)"
	@echo "  make build-windows   - 构建 Windows 平台 (amd64)"
	@echo "  make build-darwin    - 构建 macOS 平台 (amd64)"
	@echo "  make build-darwin-arm64 - 构建 macOS ARM64 平台"
	@echo "  make clean           - 清理构建产物"
	@echo "  make install         - 安装到系统路径"
	@echo "  make help            - 显示此帮助信息"

build: ## 构建当前平台的项目
	@echo "正在构建 $(BINARY_NAME)..."
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## 构建所有平台的项目
	@echo "正在构建所有平台..."
	@for platform in $(PLATFORMS); do \
		OS=$${platform%%/*}; \
		ARCH=$${platform##*/}; \
		EXT=""; \
		if [ "$$OS" = "windows" ]; then \
			EXT=".exe"; \
		fi; \
		OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH$$EXT; \
		echo "构建 $$OS/$$ARCH -> $$OUTPUT"; \
		GOOS=$$OS GOARCH=$$ARCH $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $$OUTPUT .; \
	done
	@echo "所有平台构建完成，输出目录: $(BUILD_DIR)/"

build-linux: ## 构建 Linux 平台
	@echo "正在构建 Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-windows: ## 构建 Windows 平台
	@echo "正在构建 Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

build-darwin: ## 构建 macOS 平台 (Intel)
	@echo "正在构建 macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

build-darwin-arm64: ## 构建 macOS 平台 (Apple Silicon)
	@echo "正在构建 macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

clean: ## 清理构建产物
	@echo "正在清理构建产物..."
ifeq ($(OS),Windows_NT)
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
else
	@rm -rf $(BUILD_DIR)
endif
	@echo "清理完成"

install: build ## 安装到系统路径
	@echo "正在安装 $(BINARY_NAME)..."
ifeq ($(OS),Windows_NT)
	@echo "Windows 平台请手动将 $(BUILD_DIR)/$(BINARY_NAME).exe 复制到 PATH 路径"
else
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME) 2>/dev/null || \
	 cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME) 2>/dev/null || \
	 echo "请手动将 $(BUILD_DIR)/$(BINARY_NAME) 复制到 PATH 路径"
endif
	@echo "安装完成"

