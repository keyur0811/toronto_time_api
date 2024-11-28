# Build Stage
FROM golang:1.20 as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

# Final Stage
FROM ubuntu:20.04
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]
