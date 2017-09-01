export GO15VENDOREXPERIMENT:=1
export CGO_ENABLED:=0
export GOARCH:=amd64

LOCAL_OS:=$(shell uname | tr A-Z a-z)
GOFILES:=$(shell find . -name '*.go' | grep -v -E '(./vendor)')
GOPATH_BIN:=$(shell echo ${GOPATH} | awk 'BEGIN { FS = ":" }; { print $1 }')/bin
LDFLAGS=-X github.com/StackPointCloud/digitalocean-flex-volume/pkg/version.Version=$(shell $(CURDIR)/build/git-version.sh)

all: \
	_output/bin/linux/digitalocean-flex-volume \
	_output/bin/darwin/digitalocean-flex-volume \

release: \
	clean \
	check \
	_output/release/digitalocean-flex-volume.tar.gz \

check:
	@find . -name vendor -prune -o -name '*.go' -exec gofmt -s -d {} +
	@go vet $(shell go list ./... | grep -v '/vendor/')
	@go test -v $(shell go list ./... | grep -v '/vendor/\|/e2e')

install: _output/bin/$(LOCAL_OS)/digitalocean-flex-volume
	cp $< $(GOPATH_BIN)

_output/bin/%: $(GOFILES)
	mkdir -p $(dir $@)
	GOOS=$(word 1, $(subst /, ,$*)) go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $@ github.com/StackPointCloud/digitalocean-flex-volume/cmd/$(notdir $@)

_output/release/digitalocean-flex-volume.tar.gz: _output/bin/linux/digitalocean-flex-volume _output/bin/darwin/digitalocean-flex-volume
	mkdir -p $(dir $@)
	tar czf $@ -C _output bin/linux/digitalocean-flex-volume bin/darwin/digitalocean-flex-volume

vendor:
	@glide update --strip-vendor
	@glide-vc

clean:
	rm -rf _output

.PHONY: all check clean install release vendor
