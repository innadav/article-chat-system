FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod ./go.mod
COPY go.sum ./go.sum
RUN go mod download

COPY . .

RUN go build -o /article-chat-system ./cmd/server

EXPOSE 8080

CMD ["/article-chat-system"]
