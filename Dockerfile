FROM golang:1.10.2

RUN mkdir -p /go/src/github.com/jaimemartinez88/bcg.challenge.throttler
WORKDIR /go/src/github.com/jaimemartinez88/bcg.challenge.throttler
COPY . .
RUN go build -ldflags "-linkmode external -extldflags -static" -a ./cmd/bcg.challenge.throttler/

FROM scratch
COPY --from=0 /go/src/github.com/jaimemartinez88/bcg.challenge.throttler/bcg.challenge.throttler /bcg.challenge.throttler
COPY --from=0 /go/src/github.com/jaimemartinez88/bcg.challenge.throttler/config.json /config.json
CMD ["/bcg.challenge.throttler","-config","config.json"]