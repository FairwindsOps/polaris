FROM alpine:3.20

LABEL org.opencontainers.image.authors="FairwindsOps, Inc." \
      org.opencontainers.image.vendor="FairwindsOps, Inc." \
      org.opencontainers.image.title="polaris" \
      org.opencontainers.image.description="Polaris is a cli tool to help discover deprecated apiVersions in Kubernetes" \
      org.opencontainers.image.documentation="https://polaris.docs.fairwinds.com/" \
      org.opencontainers.image.source="https://github.com/FairwindsOps/polaris" \
      org.opencontainers.image.url="https://github.com/FairwindsOps/polaris" \
      org.opencontainers.image.licenses="Apache License 2.0"

WORKDIR /usr/local/bin
RUN apk -U upgrade
RUN apk --no-cache add ca-certificates

RUN addgroup -S polaris && adduser -u 1200 -S polaris -G polaris
USER 1200
COPY polaris .

WORKDIR /opt/app

CMD ["polaris"]
