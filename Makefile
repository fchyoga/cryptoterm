APP_NAME := binance-terminal
VERSION := 1.0.0
BUILD_DIR := dist
LDFLAGS := -s -w

.PHONY: all build install run test clean cross-compile

all: build

build:
	@echo "Building $(APP_NAME) for current system..."
	go build -ldflags="$(LDFLAGS)" -o $(APP_NAME) cmd/binance-terminal/main.go
	@echo "✅ Build complete: ./$(APP_NAME)"

install: build
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	@sudo cp $(APP_NAME) /usr/local/bin/$(APP_NAME) || cp $(APP_NAME) $(HOME)/.local/bin/$(APP_NAME)
	@echo "✅ Installed successfully!"

run: build
	@./$(APP_NAME)

test:
	@echo "Running tests..."
	go test -v ./...

cross-compile:
	@echo "Cross-compiling $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	# macOS (Apple Silicon & Intel)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 cmd/binance-terminal/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 cmd/binance-terminal/main.go
	# Linux (x86_64 & ARM64)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 cmd/binance-terminal/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 cmd/binance-terminal/main.go
	# Windows (x86_64)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe cmd/binance-terminal/main.go
	@echo "✅ Cross-compile complete in $(BUILD_DIR)/:"
	@ls -lh $(BUILD_DIR)

clean:
	@rm -rf $(APP_NAME) $(BUILD_DIR)
	@echo "Cleaned build artifacts."
