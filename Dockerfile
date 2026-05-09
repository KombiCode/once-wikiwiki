FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY main.go .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o wiki .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/wiki .
VOLUME ["/storage"]
ENV PORT=80
EXPOSE 80
CMD ["./wiki"]
