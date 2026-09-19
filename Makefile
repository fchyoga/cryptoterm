APP_NAME := cryptoterm
VERSION := 1.1.0
BUILD_DIR := dist
LDFLAGS := -s -w

.PHONY: all build install run test clean cross-compile

all: build

build:
	@echo "Building $(APP_NAME) for current system..."
	go build -ldflags="$(LDFLAGS)" -o $(APP_NAME) ./cmd/cryptoterm
	@echo "✅ Build complete: ./$(APP_NAME)"

install: build
	@echo "Installing $(APP_NAME)..."
	@if [ -n "$$PREFIX" ] && [ -d "$$PREFIX/bin" ]; then \
		cp $(APP_NAME) $$PREFIX/bin/$(APP_NAME) && chmod +x $$PREFIX/bin/$(APP_NAME); \
	elif [ -w "/usr/local/bin" ]; then \
		cp $(APP_NAME) /usr/local/bin/$(APP_NAME) && chmod +x /usr/local/bin/$(APP_NAME); \
	elif command -v sudo >/dev/null 2>&1; then \
		sudo cp $(APP_NAME) /usr/local/bin/$(APP_NAME) && sudo chmod +x /usr/local/bin/$(APP_NAME); \
	else \
		mkdir -p $(HOME)/.local/bin && cp $(APP_NAME) $(HOME)/.local/bin/$(APP_NAME) && chmod +x $(HOME)/.local/bin/$(APP_NAME); \
	fi
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
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/cryptoterm
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/cryptoterm
	# Linux (x86_64 & ARM64)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/cryptoterm
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/cryptoterm
	# Windows (x86_64)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/cryptoterm
	@echo "✅ Cross-compile complete in $(BUILD_DIR)/:"
	@ls -lh $(BUILD_DIR)

clean:
	@rm -rf $(APP_NAME) $(BUILD_DIR)
	@echo "Cleaned build artifacts."
