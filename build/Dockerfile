FROM golang:1.19-alpine
WORKDIR /otus
COPY go.mod go.sum ./
RUN apk add build-base
RUN apk add curl git pkgconfig
RUN go mod download

COPY . .
RUN go build -tags go_tarantool_ssl_disable -o app ./cmd/app
EXPOSE 8081
RUN ls
CMD ["./app","-config","./cmd/app/otus.yaml"]