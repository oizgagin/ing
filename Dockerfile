FROM golang:1.20-alpine as build

ADD . /build
WORKDIR /build

RUN \
    cd cmd && go build -o /build/ing && \
    cp config.toml /build/config.toml

FROM alpine:3.17
COPY --from=build /build/ing /usr/local/bin/ing
COPY --from=build /build/config.toml /usr/local/etc/ing/config.toml

WORKDIR /usr/local/bin
ENTRYPOINT ["/usr/local/bin/ing"]
