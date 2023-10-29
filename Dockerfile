FROM golang:1.20.10 AS builder
COPY . /src
WORKDIR /src

RUN go build --race -gcflags="all=-N -l"  -o /src/server /src/cmd/main.go

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /src/server /app
COPY --from=builder /src/conf/conf.yaml /app/conf/conf.yaml

CMD ["./server", "-conf", "./conf/conf.yaml"]

