SRC := $(shell find . -type f -name '*.go')
CROSSARCH := amd64 386
CROSSOS := darwin linux openbsd netbsd freebsd windows
CROSS_SERVER := $(foreach os,$(CROSSOS),$(foreach arch,$(CROSSARCH),dist/gonzalo-server.$(os).$(arch)))

.PHONY: reset run-server cross-server

dist/gonzalo-server: $(SRC)
	@- mkdir dist 2>/dev/null
	go build -o dist/gonzalo-server ./cmd/gonzalo-server/*.go

run-server: dist/gonzalo-server
	./dist/gonzalo-server

install:
	go install github.com/frizinak/gonzalo/cmd/gonzalo-server

cross-server: $(CROSS_SERVER)

$(CROSS_SERVER): $(SRC)
	@- mkdir dist 2>/dev/null
	gox \
		-osarch="$(shell echo "$@" | cut -d'/' -f2- | cut -d'.' -f2- | sed 's/\./\//')" \
		-output="dist/gonzalo-server.{{.OS}}.{{.Arch}}" \
		./cmd/gonzalo-server/
	if [ -f "$@.exe" ]; then mv "$@.exe" "$@"; fi

reset:
	-rm -rf dist

