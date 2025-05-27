# Final binary name
BINARY_NAME=tg-game-bot-inline-button

# Path to the main file (change if different)
MAIN=main.go

# Cross-compilation variables (optional)
GOOS=linux
GOARCH=amd64

# Build the default binary
build:
 go build -o $(BINARY_NAME) $(MAIN)

# Build with inline button binary name
build-inline:
 go build -o tg-game-bot-inline-button $(MAIN)

# Build with markup button binary name
build-markup:
 go build -o tg-game-bot-markup-button $(MAIN)

# Cross-compilation build (for Linux)
build-linux:
 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME)_$(GOOS)_$(GOARCH) $(MAIN)

# Remove generated binaries
clean:
 rm -f $(BINARY_NAME) $(BINARY_NAME)_*

# Run the bot without building
run:
 go run $(MAIN)