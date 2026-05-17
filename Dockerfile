FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4 && \
    swag init -g cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/payments-api ./cmd/api


FROM alpine:3.20 AS runtime

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/payments-api /app/payments-api

EXPOSE 8080

ENTRYPOINT ["/app/payments-api"]
