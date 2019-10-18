FROM golang:1.11.4-alpine3.8 as builder

WORKDIR /go/src/producer
COPY main.go .

RUN apk add -U git
RUN go get ./...

RUN CGO_ENABLED=0 go build -a -tags netgo -o /producer 

FROM alpine:3.8
RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates
    
COPY --from=builder /producer /
ENTRYPOINT ["/producer"]
