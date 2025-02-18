FROM golang:1.22 AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker .

FROM alpine:latest
WORKDIR /root/

COPY --from=build /app/worker .
RUN chmod +x /root/worker

CMD ["./worker"]