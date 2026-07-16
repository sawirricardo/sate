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

# make ship          -> tag the next patch version and push (CI releases it)
# make ship V=v0.2.0 -> ship a specific version
ship:
	@git diff-index --quiet HEAD || { echo "commit your changes first"; exit 1; }
	@v=$(V); [ -n "$$v" ] || v=$$(git describe --tags --abbrev=0 | awk -F. -v OFS=. '{ $$3+=1; print }'); \
	git tag "$$v" && git push origin main "$$v" && echo "shipped $$v — CI is building the release"

.PHONY: build test release clean ship $(PLATFORMS)
