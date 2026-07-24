.PHONY: build wails wails-dev release clean test lint install-hooks

BINARY_NAME=tuckify

build:
	go build -o $(BINARY_NAME) .

wails:
	wails build -tags webkit2_41
	go build -o ./build/bin/tuckify .

wails-dev:
	wails dev -tags webkit2_41

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

lint:
	golangci-lint run ./...

install-hooks:
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

release: clean
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe .
