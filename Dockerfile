FROM golang:1.19 AS builder
WORKDIR /usr/local/bin/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN GOOS=linux GOARCH=arm64 go build -v -o app ./...

FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /usr/local/bin/app .
RUN ls /root

CMD ["./app"]