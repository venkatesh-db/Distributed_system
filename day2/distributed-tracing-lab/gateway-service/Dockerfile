
FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o gateway-service

EXPOSE 8080

CMD ["./gateway-service"]

