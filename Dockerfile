# syntax=docker/dockerfile:1
FROM golang:1.25-alpine

WORKDIR /go/src/app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG VERSION=local
RUN CGO_ENABLED=0 GOOS=linux go build -o etu -ldflags="-X 'main.Version=$VERSION'"

CMD ["./etu"]
