FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/generate ./cmd/generate
COPY ./cmd/certs ./cmd/certs

RUN go build -o generate cmd/generate/main.go
RUN go build -o certs cmd/certs/main.go

CMD ["echo", "Please specify a command to run"]

