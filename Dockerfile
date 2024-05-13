FROM golang:1.22.0 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

COPY web ./web

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /final .



FROM alpine:3.19

WORKDIR /usr/src/goapp

COPY --from=builder /final ./

EXPOSE 7540

CMD [ "./final" ]
