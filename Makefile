# Vulta Makefile

BINARY_NAME=vulta
INSTALL_PATH=/usr/local/bin

.PHONY: all build install clean help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) main.go

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	@echo "Done! You can now use '$(BINARY_NAME)' system-wide."

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

help:
	@echo "Usage:"
	@echo "  make build    - Build the vulta binary"
	@echo "  make install  - Install vulta to $(INSTALL_PATH)"
	@echo "  make clean    - Remove binary"
	@echo "  make help     - Show this help"
