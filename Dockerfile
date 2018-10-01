FROM golang:1.11

WORKDIR /go/app

RUN go get github.com/cespare/reflex

COPY build.sh /usr/local/bin

CMD ["reflex", "-s", "-g", "*.go", "-d", "none", "build.sh"]