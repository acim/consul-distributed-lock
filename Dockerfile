FROM golang:1.11

WORKDIR /go/app

RUN go get github.com/cespare/reflex

CMD ["reflex", "-g", "*.go", "make"]