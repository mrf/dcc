.PHONY: build run clean install

BINARY_NAME=dcc
BUILD_DIR=build

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/dcc

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)
	go clean

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)

dev:
	go run ./cmd/dcc

tidy:
	go mod tidy
