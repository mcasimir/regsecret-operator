FROM golang:1.10.1-alpine3.7
WORKDIR /go/src/github.com/mcasimir/regsecret-operator
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/mcasimir/regsecret-operator/app .
CMD ["./app"]
