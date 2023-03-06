VERSION=$(shell cat ./VERSION)
GITVERSION=$(shell git describe --tags --always)
DATE=$(shell git log --pretty=format:%cd --date=short -n1)
BRANCH=$(shell git -C . describe --tags --always --all | sed s:heads/::)
DOCKERHUB="tuggan"
OUT_DIR = build/
MKDIR_P = mkdir -p

all: directories get-depends build

get-depends:
	go get -x ./...

build:
	go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}

build-all: directories get-depends
	$(eval OS_PLATFORM_ARGS := linux windows darwin freebsd)
	$(eval OS_ARCH_ARGS := amd64)
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar; \
		done \
	done
	$(eval OS_PLATFORM_ARGS := linux windows freebsd)
	$(eval OS_ARCH_ARGS := 386)
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar; \
		done \
	done
	$(eval OS_PLATFORM_ARGS := linux windows darwin freebsd)
	$(eval OS_ARCH_ARGS := arm64)
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar; \
		done \
	done
	$(eval OS_PLATFORM_ARGS := linux windows freebsd)
	$(eval OS_ARCH_ARGS := arm)
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar; \
		done \
	done
	$(eval OS_PLATFORM_ARGS := linux)
	$(eval OS_ARCH_ARGS := mips64)
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar; \
		done \
	done
	$(eval OS_PLATFORM_ARGS := linux)
	$(eval OS_ARCH_ARGS := mips)
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_$${OS}_$${ARCH}.tar; \
		done \
	done
	$(eval OS_PLATFORM_ARGS := linux)
	$(eval OS_ARCH_ARGS := amd64 386) # arm arm64
	for OS in ${OS_PLATFORM_ARGS}; do \
		for ARCH in ${OS_ARCH_ARGS}; do \
			CGO_ENABLED=1 GOOS=$$OS GOARCH=$$ARCH go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_alpine_$${OS}_$${ARCH}; \
			tar cvf ${OUT_DIR}/goip_${VERSION}_alpine_$${OS}_$${ARCH}.tar ${OUT_DIR}/goip_${VERSION}_alpine_$${OS}_$${ARCH}; \
			gzip -9 ${OUT_DIR}/goip_${VERSION}_alpine_$${OS}_$${ARCH}.tar; \
		done \
	done

build-alpine: get-depends
	CGO_ENABLED=1 GOOS=linux go build -x -ldflags "-s -w -X main.Version=${GITVERSION} -X main.Date=${DATE} -X main.Branch=${BRANCH}" -o ${OUT_DIR}goip_${VERSION}_alpine

directories: ${OUT_DIR}

${OUT_DIR}:
	${MKDIR_P} ${OUT_DIR}

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
