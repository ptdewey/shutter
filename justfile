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
