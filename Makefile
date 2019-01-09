APP_NAME=xsync
GIT_BRANCH?=$(shell git rev-parse --verify HEAD)
VERSION=1.0.0
LDFLAGS=-ldflags "-extldflags "-static" -X main.Revision=$(GIT_BRANCH) -X main.Version=$(VERSION)"


build: clean
	CGO_ENABLED=0 GO111MODULE=on go build $(LDFLAGS) -a -o ./bin/${APP_NAME} ./cmd/xsync/

clean:
	rm -rf ./bin/*
