all: build

build:
	go get ./...
	go build .

install: build
	cp -f environator /usr/local/bin/e

unittests:
	go test github.com/dailymuse/environator/source

