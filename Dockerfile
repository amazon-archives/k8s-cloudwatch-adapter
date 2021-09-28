FROM golang:1.15-alpine as builder
RUN apk add -U --no-cache ca-certificates
WORKDIR ${GOPATH}/src/github.com/awslabs/k8s-cloudwatch-adapter
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -tags=netgo -o /adapter cmd/adapter/adapter.go

FROM busybox
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /adapter /adapter
USER 1001:1001
ENTRYPOINT ["/adapter", "--logtostderr=true"]
