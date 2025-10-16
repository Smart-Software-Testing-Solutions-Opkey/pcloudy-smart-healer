.PHONY: all build install clean rebuild

# Default target
all: build

# Build the Appium plugin
build:
	@echo "Building Appium plugin..."
	@npm run build
	@echo "Appium plugin built successfully"

# Install dependencies
install:
	@echo "Installing dependencies..."
	@npm install
	@echo "Dependencies installed successfully"

# Clean build artifacts
clean:
	@echo "Cleaning Appium plugin build artifacts..."
	@npm run clean 2>/dev/null || true
	@rm -rf lib/
	@echo "Clean completed"

# Rebuild
rebuild: clean build
