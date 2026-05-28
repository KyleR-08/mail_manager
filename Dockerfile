FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mail_manager .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mail_manager .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./mail_manager"]
