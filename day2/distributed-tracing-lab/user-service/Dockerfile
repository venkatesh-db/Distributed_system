
FROM golang:1.25

WORKDIR /app

COPY . .

RUN go build -o user-service

EXPOSE 8081

CMD ["./user-service"]