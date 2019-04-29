FROM golang:1.12.4 AS build-env
WORKDIR /go/src/github.com/reactiveops/fairwinds/

COPY . .
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 packr2 build -a -o fairwinds *.go

FROM alpine:3.9
WORKDIR /usr/local/bin
RUN apk --no-cache add ca-certificates

RUN addgroup -S fairwinds && adduser -S -G fairwinds fairwinds
USER fairwinds
COPY --from=build-env /go/src/github.com/reactiveops/fairwinds/fairwinds .

WORKDIR /opt/app

COPY --from=build-env /go/src/github.com/reactiveops/fairwinds/config.yaml ./config.yaml

CMD ["fairwinds"]
