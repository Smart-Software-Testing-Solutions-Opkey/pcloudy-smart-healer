.PHONY: all build install clean rebuild

# Default target
all: build

# Build the JavaScript client (TypeScript compilation + native addon)
build:
	@echo "Building JavaScript client..."
	@npm run build
	@echo "JavaScript client built successfully"

# Install dependencies and build native addon
install:
	@echo "Installing dependencies..."
	@npm install
	@echo "Dependencies installed successfully"

# Clean build artifacts
clean:
	@echo "Cleaning JavaScript client build artifacts..."
	@npm run clean 2>/dev/null || true
	@rm -rf build/
	@rm -rf dist/
	@echo "Clean completed"

# Rebuild
rebuild: clean build
