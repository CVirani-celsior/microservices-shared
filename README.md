# Microservices Shared Package

This package contains shared utilities, models, and middleware for the microservices architecture.

## Components

- **auth.go**: JWT authentication middleware and token management
- **models.go**: Shared data models (Order, Payment, Events)
- **pubsub.go**: Simple in-memory pub/sub messaging system (for single service use)
- **cloud_pubsub.go**: Google Cloud Pub/Sub implementation for distributed messaging
- **distributed_pubsub.go**: HTTP-based distributed pub/sub (alternative to Cloud Pub/Sub)
- **database.go**: Database initialization utilities
- **examples.go**: Usage examples for all pub/sub implementations

## Pub/Sub Options

### 1. Google Cloud Pub/Sub (Recommended for Production)
Uses Google Cloud Pub/Sub for reliable, scalable messaging between microservices.

#### Setup
1. Set environment variables:
```bash
export PUBSUB_PROJECT_ID=your-gcp-project-id
export ORDERS_TOPIC=orders-events
export PAYMENTS_TOPIC=payments-events
```

2. Initialize in your service:
```go
import shared "github.com/CVirani-celsior/microservices-shared"

func main() {
    // Initialize Cloud Pub/Sub
    if err := shared.InitCloudPubSub(); err != nil {
        log.Fatal(err)
    }
    
    // Publish an event
    ctx := context.Background()
    orderData := map[string]interface{}{
        "order_id": "123",
        "amount": 99.99,
    }
    shared.PublishOrderEvent(ctx, "order.created", orderData, "order-service")
    
    // Subscribe to events
    shared.SubscribeToPaymentEvents(ctx, "order-service-sub", handlePaymentEvent)
}
```

### 2. HTTP-based Distributed Pub/Sub
Uses HTTP calls to communicate between services.

### 3. In-Memory Pub/Sub
For single-service messaging only.

## Usage

Add this as a dependency to your microservice:

```bash
go mod init your-service-name
go get github.com/CVirani-celsior/microservices-shared
```

Then import in your service:

```go
import "github.com/CVirani-celsior/microservices-shared"
```

## Note

In a real-world scenario, you would:

1. Push this to a separate Git repository
2. Tag releases for version management
3. Use proper dependency management
4. Replace `github.com/CVirani-celsior` with your actual organization/username
