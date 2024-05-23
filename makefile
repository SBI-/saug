BINARY_NAME=saug
BUILD_DIR=bin

.PHONY: all clean linux macos-x86 macos-arm windows

all: clean linux macos-x86 macos-arm windows

clean:
	rm -rf $(BUILD_DIR)

build:
	GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) $(GOFLAGS) saug.go

linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/linux/$(BINARY_NAME) saug.go

macos-x86:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/macos-x86/$(BINARY_NAME).x86 saug.go

macos-arm:
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/macos-arm86/$(BINARY_NAME).arm saug.go

windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/windows/$(BINARY_NAME).exe saug.go
