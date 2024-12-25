FROM golang:1.22.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./app/cmd/main.go

FROM alpine:latest

WORKDIR /app

RUN mkdir config
COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yaml /app/config/
COPY --from=builder /app/fixtures /app/fixtures
COPY --from=builder /app/.env .

RUN apk add --no-cache ca-certificates

EXPOSE 8080

CMD ["./main"]
