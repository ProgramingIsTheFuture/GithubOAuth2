FROM golang:1.15

WORKDIR /go/src

ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLE=1

RUN apt-get update && \
	go get github.com/spf13/cobra/cobra && \
	go get github.com/stretchr/testify && \
	go get golang.org/x/oauth2

cmd ["tail", "-f", "/dev/null"]
