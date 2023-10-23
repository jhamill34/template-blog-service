FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/app ./cmd/app
RUN go build -o server cmd/app/main.go

COPY ./configs/app.yaml ./configs/app.yaml
COPY ./configs/rbac_model.conf ./configs/rbac_model.conf

EXPOSE 80

CMD ["./server"]

