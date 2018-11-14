FROM alpine:latest
MAINTAINER MessageBird <support@messagebird.com>

RUN apk --no-cache --update add bash ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app

ADD wabot /app/

ENTRYPOINT ["/bin/ash", "-c", "/app/wabot --http-listen-address :5000 --access-key $MB_ACCESS_KEY --public-address $PUBLIC_ADDRESS"]
