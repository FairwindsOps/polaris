FROM alpine:latest

RUN	apk add --no-cache \
  bash \
  ca-certificates \
  curl \
  wget \
  tar \
  jq

COPY get_polaris.sh ./get_polaris.sh

ENTRYPOINT ["./get_polaris.sh"]