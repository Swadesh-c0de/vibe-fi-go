BINARY_NAME=vibe
BUILD_DIR=build
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: all build clean run test install

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/vibe

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v ./...

install: build
	@mkdir -p $(INSTALL_DIR)
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)/$(BINARY_NAME)"

clean:
	rm -rf $(BUILD_DIR)
