FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/gateway ./cmd/gateway
RUN go build -o server cmd/gateway/main.go

COPY ./configs/gateway.yaml ./configs/gateway.yaml
COPY ./templates/app ./templates

EXPOSE 80

CMD ["./server"]


