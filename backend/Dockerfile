FROM golang:1.23-alpine AS builder

ARG TARGET_PATH=./gateway/cmd/server
ARG BINARY_NAME=gateway

WORKDIR /app

# Copy workspace manifest
COPY go.work ./

# Copy all go.mod files (go.sum may not exist yet — handled by go work sync)
COPY gateway/go.mod ./gateway/go.mod
COPY modules/auth/go.mod ./modules/auth/go.mod
COPY modules/users/go.mod ./modules/users/go.mod
COPY modules/vocabulary/go.mod ./modules/vocabulary/go.mod
COPY modules/notification/go.mod ./modules/notification/go.mod
COPY libs/shared/go.mod ./libs/shared/go.mod

# Download and resolve all dependencies
RUN go work sync

# Copy full source
COPY . .

# Build binary
RUN go build -o /out/${BINARY_NAME} ${TARGET_PATH}

# ---- Minimal runtime image ----
FROM alpine:3.20

ARG BINARY_NAME=gateway

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/${BINARY_NAME} /app

EXPOSE 8080

CMD ["/app"]
