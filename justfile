default: check

build:
    go build ./...

lint:
    golangci-lint run

test:
    go test ./... -race -count=1

check: lint test build

# bumps version, tags, and pushes (patch|minor|major)
release bump="patch": check
    bash scripts/release.sh {{bump}}
