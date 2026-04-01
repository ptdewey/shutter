[private]
default: build review

build:
    @pushd ./cmd/shutter && go build -o shutter ./main.go && popd

review:
    @./cmd/shutter/shutter

clean-test:
    @rm -rf ./__snapshots__
    @go test ./... -cover -coverprofile=cover.out

test:
    @go test ./... -cover -coverprofile=cover.out

cli:
    @go run cmd/cli/main.go

clean:
    @rm -rf ./__snapshots__

# Determine next version from conventional commits and tag both modules
release:
    @./scripts/version.sh

# Preview version bump without creating tags
release-dry:
    @./scripts/version.sh --dry-run
