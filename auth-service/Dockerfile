FROM golang:1.22.1 AS builder
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.20
WORKDIR /root/
COPY --from=builder /app/main .
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown root:root /root/main && chmod 755 /root/main
RUN chmod 555 /root/main

EXPOSE 4000
USER appuser
CMD ["./main"]