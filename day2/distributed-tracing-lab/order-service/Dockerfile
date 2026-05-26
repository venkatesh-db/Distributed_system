
FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o order-service

EXPOSE 8082

CMD ["./order-service"]