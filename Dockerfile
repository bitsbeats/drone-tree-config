FROM golang:1.14-alpine AS builder

RUN true \
  && apk add -U --no-cache ca-certificates git binutils

ADD . /go/src/github.com/bitsbeats/drone-tree-config
WORKDIR /go/src/github.com/bitsbeats/drone-tree-config

ENV CGO_ENABLED=0 \
    GO111MODULE=on

RUN true \
  && go test -mod=vendor ./plugin \
  && go build -mod=vendor -o drone-tree-config github.com/bitsbeats/drone-tree-config/cmd/drone-tree-config \
  && strip drone-tree-config

# ---

FROM alpine

RUN true \
  && apk add -U --no-cache ca-certificates
COPY --from=builder /go/src/github.com/bitsbeats/drone-tree-config/drone-tree-config /usr/local/bin
CMD /usr/local/bin/drone-tree-config
