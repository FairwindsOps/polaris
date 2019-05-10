FROM golang:1.12.4 AS build-env
WORKDIR /go/src/github.com/reactiveops/polaris/

COPY . .
RUN go get -u github.com/gobuffalo/packr/v2/packr2
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 packr2 build -a -o polaris *.go

FROM alpine:3.9
WORKDIR /usr/local/bin
RUN apk --no-cache add ca-certificates

RUN addgroup -S polaris && adduser -S -G polaris polaris
USER polaris
COPY --from=build-env /go/src/github.com/reactiveops/polaris/polaris .

WORKDIR /opt/app

COPY --from=build-env /go/src/github.com/reactiveops/polaris/config.yaml ./config.yaml

CMD ["polaris"]
