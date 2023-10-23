FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/auth ./cmd/auth
RUN go build -o server cmd/auth/main.go

COPY ./configs/auth.yaml ./configs/auth.yaml
COPY ./configs/rbac_model.conf ./configs/rbac_model.conf
COPY ./templates/auth ./templates

COPY ./keys/public.pem ./keys/public.pem
COPY ./keys/private.pem ./keys/private.pem

EXPOSE 80

CMD ["./server"]

