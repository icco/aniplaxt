FROM golang:1.17-alpine as builder

ENV GOPROXY="https://proxy.golang.org"
ENV GO111MODULE="on"
ENV NAT_ENV="production"

EXPOSE 8080

WORKDIR /go/src/github.com/icco/aniplaxt/

RUN apk add --no-cache git
COPY . .

RUN go build -v -o /go/bin/server ./server

CMD ["/go/bin/server"]
