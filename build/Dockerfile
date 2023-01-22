FROM golang:1.19-alpine
WORKDIR /otus
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o app ./cmd/app
EXPOSE 8081
CMD ["./app","-config","./cmd/app/otus.yaml"]