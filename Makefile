BINARY := spacehoggers
VERSION := 2019-01-07
SOURCES := $(wildcard *.go)
COMMIT_ID := $(shell git describe --tags --always)
BUILD_TIME := $(shell date +%FT%T%:z)
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.BUILD_DATE=${BUILD_TIME} -X main.COMMIT_ID=${COMMIT_ID} -d -s -w"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
		env CGO_ENABLED=0 go build ${LDFLAGS} -o $@ ${SOURCES}

.PHONY: install
install:
		env CGO_ENABLED=0 go install ${LDFLAGS} ./...

.PHONY: clean
clean:
		if [ -f ${BINARY} ]; then rm -f ${BINARY}; fi

