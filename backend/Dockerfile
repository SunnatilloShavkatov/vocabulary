FROM golang:1.23-alpine

WORKDIR /app

COPY go.work ./
COPY gateway/go.mod ./gateway/go.mod
COPY modules/auth/go.mod ./modules/auth/go.mod
COPY modules/vocabulary/go.mod ./modules/vocabulary/go.mod
COPY libs/shared/go.mod ./libs/shared/go.mod
RUN go work sync

COPY . .

EXPOSE 8080

CMD ["go", "run", "./gateway/cmd/server"]

