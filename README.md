# Microservices Shared Package

This package contains shared utilities, models, and middleware for the microservices architecture.

## Components

- **auth.go**: JWT authentication middleware and token management
- **models.go**: Shared data models (Order, Payment, Events)
- **pubsub.go**: Simple pub/sub messaging system
- **database.go**: Database initialization utilities

## Usage

Add this as a dependency to your microservice:

```bash
go mod init your-service-name
go get github.com/yourorg/microservices-shared
```

Then import in your service:

```go
import "github.com/yourorg/microservices-shared"
```

## Note

In a real-world scenario, you would:

1. Push this to a separate Git repository
2. Tag releases for version management
3. Use proper dependency management
4. Replace `github.com/yourorg` with your actual organization/username
