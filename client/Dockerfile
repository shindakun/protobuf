# builder image
FROM golang:latest as builder
RUN mkdir /build
ADD *.* /build/
WORKDIR /build
RUN go mod init github.com/shindakun/client
RUN go get ./...
RUN CGO_ENABLED=0 go build -a -o ftp-load-client .

# generate final image
FROM alpine:latest
COPY --from=builder /build/ftp-load-client /usr/local/bin/ftp-load-client
RUN chmod +x /usr/local/bin/ftp-load-client
# executable
ENTRYPOINT [ "ftp-load-client" ]
# arguments
CMD [ "34.82.64.125" ]