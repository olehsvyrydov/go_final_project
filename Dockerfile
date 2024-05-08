FROM golang:1.22.0 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

COPY web ./web

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /final

ENV TODO_PORT=7540

ENV TODO_DBFILE=./sheduler.db

ENTRYPOINT [ "/final" ]



# FROM alpine:3.19

# WORKDIR /usr/src/goapp

# COPY --from=builder /final ./

# ENV TODO_PORT=7540

# ENV TODO_DBFILE=./sheduler.db

# EXPOSE 7540

# ENTRYPOINT [ "sh", "./final" ]
