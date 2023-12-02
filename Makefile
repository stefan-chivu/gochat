VERSION := "$(shell git describe --tags)-$(shell git rev-parse --short HEAD)"
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS += -X github.com/stefan-chivu/gochat/gochat.Version=$(VERSION)
GOLDFLAGS += -X github.com/stefan-chivu/gochat/gochat.Buildtime=$(BUILDTIME)

GOFLAGS = -ldflags "$(GOLDFLAGS)"

.PHONY: build release

build: clean
	go build -o gochat-app $(GOFLAGS) ./gochat
	chmod +x gochat-app
	./gochat-app -version

clean:
	rm -f gochat-app
	rm -f cover.out
	rm -f cpu.pprof

cover:
	go test -count=1 -cover -coverprofile=cover.out ./...
	go tool cover -func=cover.out

debug: build
	./gochat-app -PProf -CPUProfile=cpu.pprof -ServerTLSCert=server.crt -ServerTLSKey=server.key

lint:
	go fmt ./ ./gochat/...
	go vet

release:
	mkdir -p release
	rm -f release/gochat-app release/gochat-app.exe
ifeq ($(shell go env GOOS), windows)
	go build -o release/gochat-app.exe $(GOFLAGS) ./gochat
	cd release; zip -m "gochat-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip" gochat-app.exe
else
	go build -o release/gochat-app $(GOFLAGS) ./gochat
	cd release; zip -m "gochat-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip" gochat-app
endif
	cd release; sha256sum "gochat-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip" > "gochat-app-$(shell git describe --tags --abbrev=0)-$(shell go env GOOS)-$(shell go env GOARCH).zip.sha256"


run: build
	./gochat-app -ServerTLSCert=server.crt -ServerTLSKey=server.key

sync:
	go get ./...

test: clean
	go test -count=1 -cover ./...

tls:
	openssl ecparam -genkey -name secp384r1 -out server.key
	openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 -subj "/CN=selfsigned.gochat.local"

update:
	go mod tidy
	go get -u ./...
