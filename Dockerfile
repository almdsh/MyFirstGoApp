FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod downland

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

EXPOSE 8080

CMD ["./main"]