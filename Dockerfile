# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o wiki .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

COPY --from=builder /app/wiki .

VOLUME ["/storage"]

ENV PORT=8080
EXPOSE 8080

CMD ["./wiki"]
