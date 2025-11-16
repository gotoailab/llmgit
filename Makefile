.PHONY: build clean install help

# 构建目录
BUILD_DIR := build
BINARY_NAME := llmgit

# Go 参数
GO := go
GOFLAGS := -v

help: ## 显示帮助信息
	@echo "可用的命令:"
	@echo "  make build     - 构建项目到 build 目录"
	@echo "  make clean     - 清理构建产物"
	@echo "  make install   - 安装到系统路径"
	@echo "  make help      - 显示此帮助信息"

build: ## 构建项目
	@echo "正在构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## 清理构建产物
	@echo "正在清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@echo "清理完成"

install: build ## 安装到系统路径
	@echo "正在安装 $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME) || \
	 cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME) || \
	 echo "请手动将 $(BUILD_DIR)/$(BINARY_NAME) 复制到 PATH 路径"
	@echo "安装完成"

