FROM scratch

COPY $BINARY bcg.challenge.throttler
COPY $CONFIG config.json

CMD ["bcg.challenge.throttler","-config", "config.json"]