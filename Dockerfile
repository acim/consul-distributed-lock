FROM golang:1.11

WORKDIR /go/app
COPY . .

RUN go get github.com/cespare/reflex

CMD ["reflex","-c","/go/app/reflex.conf"]