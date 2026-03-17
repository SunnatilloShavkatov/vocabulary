FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy workspace manifest
COPY go.work ./

# Copy all go.mod files (go.sum may not exist yet — handled by go work sync)
COPY gateway/go.mod ./gateway/go.mod
COPY modules/auth/go.mod ./modules/auth/go.mod
COPY modules/vocabulary/go.mod ./modules/vocabulary/go.mod
COPY libs/shared/go.mod ./libs/shared/go.mod

# Download and resolve all dependencies
RUN go work sync

# Copy full source
COPY . .

# Build binary
RUN go build -o /gateway ./gateway/cmd/server

# ---- Minimal runtime image ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /gateway /gateway

EXPOSE 8080

CMD ["/gateway"]
