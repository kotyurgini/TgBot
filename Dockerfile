FROM golang:alpine3.21

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/.

CMD ["./main", "--config-path", "./data/prod.yaml"]