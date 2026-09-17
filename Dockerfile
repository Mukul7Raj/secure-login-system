FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install gcc and musl-dev just in case, though modernc.org/sqlite doesn't strictly need it, some dependencies might.
RUN apk add --no-cache gcc musl-dev

COPY go.mod ./
# No go.sum yet, so let's let Go resolve dependencies
# We run go mod tidy to generate go.sum
COPY . .
RUN go mod tidy
RUN go build -o secure-cli .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/secure-cli /app/secure-cli

# Make sure data directory exists
RUN mkdir -p /app/data

# Run the CLI
CMD ["/app/secure-cli"]
