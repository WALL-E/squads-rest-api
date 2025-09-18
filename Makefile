APP_NAME=squads-rest-api
SERVER_BIN=server
TEST_BIN=test-api
SETUP_BIN=setup-db

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS = -X 'main.Version=$(VERSION)' \
          -X 'main.BuildTime=$(BUILD_TIME)' \
          -X 'main.GitCommit=$(GIT_COMMIT)'

# 运行服务器
run:
	go mod tidy
	go run -ldflags "$(LDFLAGS)" ./cmd/server

# 运行API测试
test:
	go mod tidy
	go run -ldflags "$(LDFLAGS)" ./cmd/test

# 运行数据库设置
setup:
	go mod tidy
	go run -ldflags "$(LDFLAGS)" ./cmd/setup

# 构建所有组件
build: build-server build-test build-setup

# 构建服务器
build-server:
	go mod tidy
	go build -ldflags "$(LDFLAGS)" -o $(SERVER_BIN) ./cmd/server

# 构建测试工具
build-test:
	go mod tidy
	go build -ldflags "$(LDFLAGS)" -o $(TEST_BIN) ./cmd/test

# 构建数据库设置工具
build-setup:
	go mod tidy
	go build -ldflags "$(LDFLAGS)" -o $(SETUP_BIN) ./cmd/setup

# 清理构建文件
clean:
	rm -f $(SERVER_BIN) $(TEST_BIN) $(SETUP_BIN) app.db

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  run         - 运行API服务器"
	@echo "  test        - 运行API测试"
	@echo "  setup       - 运行数据库设置"
	@echo "  build       - 构建所有组件"
	@echo "  build-server - 构建服务器"
	@echo "  build-test  - 构建测试工具"
	@echo "  build-setup - 构建数据库设置工具"
	@echo "  clean       - 清理构建文件"
	@echo "  help        - 显示此帮助信息"

.PHONY: run test setup build build-server build-test build-setup clean help
