FROM golang:1.11 AS build-env
WORKDIR /go/src/github.com/reactiveops/fairwinds/

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o fairwinds *.go

FROM alpine:3.8
WORKDIR /usr/local/bin
RUN apk --no-cache add ca-certificates

USER nobody
COPY --from=build-env /go/src/github.com/reactiveops/fairwinds/fairwinds .

WORKDIR /opt/app

ENTRYPOINT ["fairwinds"]
