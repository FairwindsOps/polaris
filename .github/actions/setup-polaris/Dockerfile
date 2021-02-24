# The action uses an own Dockerfile on purpose because the root Dockerfile takes way too long to build for an action

FROM alpine:3.10

RUN	apk add --no-cache \
  bash \
  ca-certificates \
  curl \
  wget \
  tar \
  jq

COPY get_polaris.sh /get_polaris.sh

ENTRYPOINT ["/get_polaris.sh"]
