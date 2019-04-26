FROM golang:1.11.4 AS build-env
WORKDIR /go/src/github.com/reactiveops/fairwinds/

COPY . .
RUN go get -u github.com/gobuffalo/packr/packr
RUN packr
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o fairwinds *.go

FROM alpine:3.8
WORKDIR /usr/local/bin
RUN apk --no-cache add ca-certificates

RUN addgroup -S fairwinds && adduser -S -G fairwinds fairwinds
USER fairwinds
COPY --from=build-env /go/src/github.com/reactiveops/fairwinds/fairwinds .

WORKDIR /opt/app

COPY --from=build-env /go/src/github.com/reactiveops/fairwinds/config.yaml ./config.yaml

CMD ["fairwinds"]
