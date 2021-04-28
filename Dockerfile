# Build the manager binary
FROM golang:1.15 as builder

WORKDIR /go/src/github.com/zshi-redhat/network-device-operator
COPY . .

# Build
RUN make build

FROM centos:8
WORKDIR /
COPY --from=builder /go/src/github.com/zshi-redhat/network-device-operator/bin/manager /usr/bin/manager
COPY --from=builder /go/src/github.com/zshi-redhat/network-device-operator/bin/network-device-daemon /usr/bin/network-device-daemon
COPY bindata /bindata
USER 65532:65532

CMD ["/usr/bin/manager"]
