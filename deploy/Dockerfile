FROM alpine:3.9 as alpine

RUN apk add -U --no-cache ca-certificates

FROM busybox
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY adapter /
USER 1001:1001
ENTRYPOINT ["/adapter", "--logtostderr=true"]
