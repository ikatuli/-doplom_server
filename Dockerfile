# syntax=docker/dockerfile:1

# docker build --tag doplom_server ./
FROM golang:1.16-alpine

WORKDIR /app
COPY main.go ./
COPY go.mod ./
COPY static/ ./static
COPY user/ ./user
COPY config.toml ./
RUN go mod download
RUN go mod tidy
RUN go build main.go

EXPOSE  9000

CMD ["/app/main"]

# docker run -p 9000:9000 doplom_server
