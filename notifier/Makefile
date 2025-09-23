# Variables
IMAGE_NAME = tg-notifier
CONTAINER_NAME = notifier
DATA_DIR = $(PWD)/../data

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[1;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: help build run stop remove restart logs status clean

# Help command
help:
    @echo "$(BLUE)TG Echo Post Bot Notifier - Docker Commands$(NC)"
    @echo ""
    @echo "$(GREEN)Available commands:$(NC)"
    @echo "  $(YELLOW)build$(NC)    - Build Docker image"
    @echo "  $(YELLOW)run$(NC)      - Run container with cron"
    @echo "  $(YELLOW)stop$(NC)     - Stop running container"
    @echo "  $(YELLOW)remove$(NC)   - Remove container"
    @echo "  $(YELLOW)restart$(NC)  - Stop, remove, build and run (full restart)"
    @echo "  $(YELLOW)logs$(NC)     - Show container logs in real-time"
    @echo "  $(YELLOW)status$(NC)   - Show container status"
    @echo "  $(YELLOW)clean$(NC)    - Remove image and container"
    @echo ""
    @echo "$(GREEN)Data directory:$(NC) $(DATA_DIR)"

# Build Docker image
build:
    @echo "$(BLUE)Building Docker image...$(NC)"
    @if [ ! -f ".env" ]; then \
        echo "$(RED)Error: .env file not found in current directory!$(NC)"; \
        echo "Please create .env file in notifier/ directory first."; \
        exit 1; \
    fi
    @docker build -t $(IMAGE_NAME) .
    @echo "$(GREEN)Docker image built successfully!$(NC)"

# Run container
run: check-data-dir
    @echo "$(BLUE)Starting container...$(NC)"
    @docker run -d \
        --name $(CONTAINER_NAME) \
        --volume $(DATA_DIR):/app/data \
        --env-file .env \
        --restart unless-stopped \
        $(IMAGE_NAME)
    @echo "$(GREEN)Container started successfully!$(NC)"
    @echo "Use '$(YELLOW)make logs$(NC)' to view logs in real-time"
    @echo "Use '$(YELLOW)make status$(NC)' to check container status"

# Stop container
stop:
    @echo "$(BLUE)Stopping container...$(NC)"
    @docker stop $(CONTAINER_NAME) 2>/dev/null || echo "$(YELLOW)Container was not running$(NC)"
    @echo "$(GREEN)Container stopped!$(NC)"

# Remove container
remove:
    @echo "$(BLUE)Removing container...$(NC)"
    @docker rm $(CONTAINER_NAME) 2>/dev/null || echo "$(YELLOW)Container was not found$(NC)"
    @echo "$(GREEN)Container removed!$(NC)"

# Full restart
restart: stop remove build run
    @echo "$(GREEN)Full restart completed!$(NC)"

# View logs in real-time
logs:
    @echo "$(BLUE)Showing logs... (Press Ctrl+C to exit)$(NC)"
    @docker logs -f $(CONTAINER_NAME) 2>/dev/null || echo "$(RED)Container not found or not running$(NC)"

# Container status
status:
    @echo "$(BLUE)Container status:$(NC)"
    @docker ps -a --filter name=$(CONTAINER_NAME) --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}" || echo "$(RED)Docker not available$(NC)"

# Check data directory
check-data-dir:
    @echo "$(BLUE)Checking data directory...$(NC)"
    @if [ ! -d "$(DATA_DIR)" ]; then \
        echo "$(YELLOW)Creating data directory: $(DATA_DIR)$(NC)"; \
        mkdir -p $(DATA_DIR); \
    fi
    @echo "$(GREEN)Data directory ready: $(DATA_DIR)$(NC)"

# Cleanup (remove image and container)
clean: stop remove
    @echo "$(BLUE)Removing Docker image...$(NC)"
    @docker rmi $(IMAGE_NAME) 2>/dev/null || echo "$(YELLOW)Image was not found$(NC)"
    @echo "$(GREEN)Cleanup completed!$(NC)"

# Enter the container for debugging
shell:
    @echo "$(BLUE)Entering container shell...$(NC)"
    @docker exec -it $(CONTAINER_NAME) /bin/sh || echo "$(RED)Container not running$(NC)"

# View cron logs inside the container
cron-logs:
    @echo "$(BLUE)Showing cron logs...$(NC)"
    @docker exec $(CONTAINER_NAME) cat /var/log/cron/notifier.log 2>/dev/null || echo "$(RED)No cron logs found$(NC)"
