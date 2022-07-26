FROM golang:1.18 as builder
WORKDIR /go/src
COPY . .
RUN make build

FROM alpine
COPY --from=builder /go/src/certs ./certs
COPY --from=builder /go/src/templates ./templates
COPY --from=builder /go/src/bin/kafka-producer /usr/bin
ENTRYPOINT [ "kafka-producer" ]