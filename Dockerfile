FROM golang:1.11

WORKDIR /go/app

RUN go get github.com/cespare/reflex

COPY build.sh /usr/local/bin

CMD ["reflex", "-g", "*.go", "build.sh"]