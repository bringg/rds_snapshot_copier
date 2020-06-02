FROM golang:1.14 as builder
ENV CGO_ENABLED=0

WORKDIR /src

# get dependencies
COPY go.mod go.sum ./
RUN go mod download

# compile
COPY . .
RUN go install

FROM busybox
LABEL maintainer "Alex Simenduev <shamil@bringg.com>"

USER nobody
ENTRYPOINT ["rds_snapshot_copier"]
CMD ["-help"]

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /go/bin/rds_snapshot_copier /usr/local/bin/
