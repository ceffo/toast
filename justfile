default: check

build:
    go build ./...

lint:
    golangci-lint run

test:
    go test ./... -race -count=1

check: lint test build

release version: check
    git tag -a "v{{version}}" -m "Release v{{version}}"
    git push origin "v{{version}}"
