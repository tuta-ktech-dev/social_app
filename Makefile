APP_NAME=social-app

# Rebuild Docker container
build:
	docker compose build

# Start the project
up:
	docker compose up

# Start the project (rebuild first)
up-build:
	docker compose up --build

# Stop all services
down:
	docker compose down

# Remove containers, networks, volumes
clean:
	docker compose down -v

# View Go app logs
logs:
	docker logs go-chat-app

# View Redis logs
logs-redis:
	docker logs redis-server

# Enter app container shell
shell:
	docker exec -it go-chat-app /bin/bash

# Enter Redis CLI in container
redis-cli:
	docker exec -it redis-server redis-cli
