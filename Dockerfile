FROM golang:1.11

WORKDIR /go/app
COPY . .

RUN CGO_ENABLED=0 go build -installsuffix cgo -ldflags "-s -w" -o /go/bin/app .

CMD /go/bin/app