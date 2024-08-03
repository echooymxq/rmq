# linux
.PHONY: linux-build
linux-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(CURDIR)/bin/rmq cmd/main.go


# darwin
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(CURDIR)/bin/rmq cmd/main.go