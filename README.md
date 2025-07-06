# Social App

A basic social network backend project built with Go and Redis, containerized using Docker Compose.

## Features
- User can interact with posts (like, comment, share)
- Real-time messaging between users
- Post sharing functionality
- Redis used for caching and real-time features
- Easily extensible for more social features

## Project Structure
```
social_app/
  app/                      # Go application source code
    config/                 # Configuration files
      redis.go              # Redis connection setup
    internal/               # Internal application code
      domain/               # Domain models and interfaces
        user_status.go      # User status domain logic
      repository/           # Data access layer
        redis_user_status.go # Redis implementation for user status
    Dockerfile              # Build Go app image
    main.go                 # Main entry point
    go.mod, go.sum          # Go dependencies
  docs/                     # Documentation
    users/                  # User-related features docs
      online-status/        # User online status documentation
  redis.conf                # Redis configuration
  docker-compose.yml        # Multi-service orchestration
  Makefile                  # Automation commands
```

## Prerequisites
- Docker & Docker Compose installed
- (Optional) Go 1.21+ for local development

## Quick Start
```bash
git clone <repo-url>
cd social_app
# Build and start all services
docker compose up --build
```
- The Go app will be available in the `go-social-app` container.
- Redis runs in the `redis-server` container (port 6379).

## Useful Commands
- `make build`      : Build Docker containers
- `make up`         : Start all services
- `make up-build`   : Rebuild and start all services
- `make down`       : Stop all services
- `make clean`      : Remove containers, networks, volumes
- `make logs`       : View Go app logs
- `make logs-redis` : View Redis logs
- `make shell`      : Enter Go app container shell
- `make redis-cli`  : Enter Redis CLI

## Extending the Project
- Add more features in `app/main.go` or split into multiple Go files/packages.
- Update `docker-compose.yml` to add new services (e.g., database, frontend).
- Adjust `redis.conf` for advanced Redis settings.

## License
MIT 