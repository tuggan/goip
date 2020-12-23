VERSION=$(shell cat ./VERSION)
GITVERSION=$(shell git describe --tags --always)
DATE=$(shell git log --pretty=format:%cd --date=short -n1)
BRANCH=$(shell git -C . describe --tags --always --all | sed s:heads/::)
DOCKERHUB="tuggan"

all: get-depends build

get-depends:
	go get -x ./...

build:
	go build -x -ldflags "-X main.version=${GITVERSION} -X main.date=${DATE} -X main.branch=${BRANCH}"

build-alpine: get-depends
	CGO_ENABLED=1 GOOS=linux go build -x -ldflags "-X main.version=${GITVERSION} -X main.date=${DATE} -X main.branch=${BRANCH}"

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
	#goreleaser --rm-dist

copyconfig:
ifeq (,$(wildcard ./goip.toml))
	cp config/goip.toml ./
endif

dockerimage: copyconfig
	docker image build -t ${DOCKERHUB}/goip:${GITVERSION} -t ${DOCKERHUB}/goip:dev .

.PHONY: build install test fmt clean release
