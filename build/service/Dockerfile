FROM golang:1.21-alpine
ARG SERVICE

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/${SERVICE} ./cmd/${SERVICE}
RUN go build -o service cmd/${SERVICE}/main.go

COPY ./templates ./templates
COPY ./migrations ./migrations

CMD ["./service"]

