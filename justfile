default: check

build:
    go build ./...

lint:
    golangci-lint run

test:
    go test ./... -race -count=1

check: lint test build
