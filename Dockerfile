FROM golang:1.18 as builder
WORKDIR /go/src
COPY . .
RUN make build

FROM alpine
COPY --from=builder /go/src/.env .
COPY --from=builder /go/src/server.* ./
COPY --from=builder /go/src/templates ./templates
COPY --from=builder /go/src/bin/kafka-producer /usr/bin
ENTRYPOINT [ "kafka-producer" ]