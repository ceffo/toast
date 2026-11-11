default: check

build:
    go build ./...

lint:
    golangci-lint run

test:
    go test ./... -race -count=1

# shell tests of the release scripts (docs/versioning.md); need git and jq
test-scripts:
    bash scripts/test-release-version.sh
    bash scripts/test-version-tag.sh
    bash scripts/test-backfill-releases.sh
    bash scripts/test-check-pr-title.sh

check: lint test build test-scripts
