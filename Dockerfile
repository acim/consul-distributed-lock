FROM golang:1.11

VOLUME /go/app

RUN go get github.com/cespare/reflex

CMD ["reflex","-c","/go/app/reflex.conf"]