build:
	go build

test:
	go test -v ./...

install:
	cp ./unit ~/.local/bin/unit