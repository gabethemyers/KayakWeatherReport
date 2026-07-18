FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o kayak-api ./

#Runner
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/kayak-api .

EXPOSE 8080

CMD ["./kayak-api"]
