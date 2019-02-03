VERSION := $(shell cat ./VERSION)

all: build

build:
	go build -x

install:
	go install -v

test:
	go test -v ./...

fmt:
	go fmt -x ./...

clean:
	go clean -x 

release:
	git tag -a $(VERSION) -m "Release" || true
	git push origin $(VERSION)
	goreleaser --rm-dist

.PHONY: build install test fmt clean release
