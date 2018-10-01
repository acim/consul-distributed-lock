build:
	CGO_ENABLED=0 go build -installsuffix cgo -ldflags '-s -w' -o /go/bin/app
	/go/bin/app