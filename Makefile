BINARY    := sate
DIST      := dist
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

build:
	go build -o $(BINARY) .

test:
	go test ./...

release: test $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
	go build -trimpath -ldflags="-s -w" \
	-o $(DIST)/$(BINARY)-$(subst /,-,$@)$(if $(findstring windows,$@),.exe) .

clean:
	rm -rf $(DIST) $(BINARY)

.PHONY: build test release clean $(PLATFORMS)
