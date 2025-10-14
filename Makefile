# # Final binary name
# BINARY_NAME=tg-game-bot

# # Path to the main file (change if different)
# MAIN=main.go

# # Cross-compilation variables (optional)
# GOOS=linux
# GOARCH=amd64

# # Build the default binary
# build:
# 	go build -o $(BINARY_NAME) $(MAIN)

# # Build with inline button binary name
# build-inline:
# 	go build -o tg-game-bot-inline-button $(MAIN)

# # Build with markup button binary name
# build-markup:
# 	go build -o tg-game-bot-markup-button $(MAIN)

# # Cross-compilation build (for Linux)
# build-linux:
# 	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME)_$(GOOS)_$(GOARCH) $(MAIN)

# # Remove generated binaries
# clean:
# 	rm -f $(BINARY_NAME) $(BINARY_NAME)_*

# # Run the bot without building
# run:
# 	go run $(MAIN)

# Final binary name
BINARY_NAME=tg-game-bot
MAIN=main.go

# Cross-compilation variables (optional)
GOOS=linux
GOARCH=amd64

# Docker variables
DOCKER_IMAGE=tg-game-bot
DOCKER_CONTAINER=tg-game-bot
DATA_DIR=$(PWD)/data

# ======================================================
# ================ Local build/run =====================
# ======================================================

build:
	go build -o $(BINARY_NAME) $(MAIN)

build-inline:
	go build -o tg-game-bot-inline-button $(MAIN)

build-markup:
	go build -o tg-game-bot-markup-button $(MAIN)

build-linux:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME)_$(GOOS)_$(GOARCH) $(MAIN)

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)_*

run:
	go run $(MAIN)

# ======================================================
# ================ Docker build/run ====================
# ======================================================

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
# 	docker run -d \
# 		--name $(DOCKER_CONTAINER) \
# 		-e TELEGRAM_TOKEN="$(TELEGRAM_TOKEN)" \
# 		-v $(DATA_DIR):/app/data \
# 		$(DOCKER_IMAGE)

	docker run -d \
		--name $(DOCKER_CONTAINER) \
		--env-file .env \
		-v $(DATA_DIR):/app/data \
		$(DOCKER_IMAGE)

container-stop:
	docker stop $(DOCKER_CONTAINER) || true
	docker rm $(DOCKER_CONTAINER) || true

docker-logs:
	docker logs -f $(DOCKER_CONTAINER)

docker-clean: docker-stop
	docker rmi $(DOCKER_IMAGE) || true

container-start:
	make docker-build || true
	make docker-run || true
	docker logs -f $(DOCKER_CONTAINER) || true

container-log:
	docker logs -f $(DOCKER_CONTAINER)

container-status:
	docker ps -a