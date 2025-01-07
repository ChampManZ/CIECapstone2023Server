FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o capstone

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/capstone .

COPY html/ /app/html/

EXPOSE 8443

CMD ["./capstone"]
