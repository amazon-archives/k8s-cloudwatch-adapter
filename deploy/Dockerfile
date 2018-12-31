FROM alpine:3.6 as alpine

RUN apk add -U --no-cache ca-certificates

FROM BASEIMAGE
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY adapter /
USER 1001:1001
ENTRYPOINT ["/adapter", "--logtostderr=true"]
