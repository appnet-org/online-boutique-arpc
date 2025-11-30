# Build stage
FROM golang:1.24.0 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy only the necessary Go mod files to cache dependencies
COPY go.mod go.sum ./
# Copy proto directory since it's referenced in go.mod replace directive
COPY proto/ proto/
# Copy arpc-quic dependency (referenced by replace directive)
COPY arpc-quic/ arpc-quic/

# Download and cache Go dependencies
RUN go mod download

# Copy the entire project directory to the container
COPY . .

# Build the Go application with optimized flags
RUN go build -ldflags="-s -w" -o /app/onlineboutique ./cmd/...

# Final stage
FROM alpine:latest
# FROM ubuntu:20.04
RUN apk add gcompat

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/onlineboutique .
COPY --from=builder /app/services/templates /app/templates
COPY --from=builder /app/services/static /app/static
COPY --from=builder /app/services/data /app/data
COPY --from=builder /app/services/tracing /app/tracing

RUN chmod +x /app/onlineboutique

# Set environment variables
ENV CART_SERVICE_ADDR="cart:11001" \
    CART_REDIS_ADDR="cart-redis:6379" \
    PRODUCT_CATALOG_SERVICE_ADDR="productcatalog:11002" \
    CURRENCY_SERVICE_ADDR="currency:11003" \
    PAYMENT_SERVICE_ADDR="payment:11004" \
    SHIPPING_SERVICE_ADDR="shipping:11005" \
    EMAIL_SERVICE_ADDR="email:11006" \
    CHECKOUT_SERVICE_ADDR="checkout:11007" \
    RECOMMENDATION_SERVICE_ADDR="recommendation:11008" \
    AD_SERVICE_ADDR="ad:11009" \
    SHOPPING_ASSISTANT_SERVICE_ADDR="shoppingassistant:80"
