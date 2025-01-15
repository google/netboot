# netboot/Makefile

THIS := $(abspath $(lastword $(MAKEFILE_LIST)))
HERE := $(patsubst %/,%,$(dir $(THIS)))

GOCMD:=go

# Local customizations to the above.
ifneq ($(wildcard Makefile.defaults),)
include Makefile.defaults
endif

all:
	$(error Please request a specific thing, there is no default target)

.PHONY: build
build: update-ipxe
	mkdir -p build/
	$(GOCMD) build ./cmd/pixiecore -o build/

.PHONY: test
test:
	$(GOCMD) test ./...
	$(GOCMD) test -race ./...

.PHONY: lint
lint:
	$(GOCMD) tool vet .

.PHONY: update-ipxe
update-ipxe:
	$(MAKE) -C third_party/ipxe/src \
	EMBED=$(HERE)/pixiecore/boot.ipxe \
	bin/ipxe.pxe \
	bin/undionly.kpxe \
	bin-x86_64-efi/ipxe.efi \
	bin-i386-efi/ipxe.efi

	mkdir -p artifacts/ipxe/bin artifacts/ipxe/bin-x86_64-efi artifacts/ipxe/bin-i386-efi

	cp third_party/ipxe/src/bin/ipxe.pxe artifacts/ipxe/bin/
	cp third_party/ipxe/src/bin/undionly.kpxe artifacts/ipxe/bin/
	cp third_party/ipxe/src/bin-x86_64-efi/ipxe.efi artifacts/ipxe/bin-x86_64-efi/
	cp third_party/ipxe/src/bin-i386-efi/ipxe.efi artifacts/ipxe/bin-i386-efi/


