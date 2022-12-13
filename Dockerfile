FROM alpine:3.17
WORKDIR /usr/local/bin
RUN apk -U upgrade
RUN apk --no-cache add ca-certificates

RUN addgroup -S polaris && adduser -u 1200 -S polaris -G polaris
USER 1200
COPY polaris .

WORKDIR /opt/app

CMD ["polaris"]
