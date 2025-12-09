FROM golang:1.25 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Production stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy the binary
COPY --from=builder /app/main .

# Copy config (needed for default values)
COPY --from=builder /app/config ./config

EXPOSE 3000

USER nonroot:nonroot

CMD ["./main", "serve"]
