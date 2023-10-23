FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/migrator ./cmd/migrator
RUN go build -o migrator cmd/migrator/main.go

COPY ./configs/migrator.yaml ./configs/migrator.yaml
COPY ./migrations ./migrations

CMD ["./migrator"]

