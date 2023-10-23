FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./internal ./internal
COPY ./cmd/mail ./cmd/mail
RUN go build -o server cmd/mail/main.go

COPY ./configs/mail.yaml ./configs/mail.yaml
COPY ./keys/dkim-private.pem ./keys/dkim-private.pem
COPY ./keys/dkim-public.pem ./keys/dkim-public.pem

EXPOSE 25

CMD ["./server"]

