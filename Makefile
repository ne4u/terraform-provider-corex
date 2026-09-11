VERSION ?= dev
BINARY = terraform-provider-corex

.PHONY: build install test vet fmt clean

build:
	go build -o $(BINARY) .

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/ne4u/corex/$(VERSION)/darwin_arm64
	cp $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/ne4u/corex/$(VERSION)/darwin_arm64/

test:
	go test ./... -v

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -f $(BINARY)
