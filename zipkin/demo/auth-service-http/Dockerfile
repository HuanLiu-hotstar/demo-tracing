FROM golang:1.16

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build server.go

ENTRYPOINT ["/app/server"]
