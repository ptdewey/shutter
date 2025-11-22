clean-test:
    @rm -rf ./__snapshots__
    @go test ./... -cover -coverprofile=cover.out

test:
    @go test ./... -cover -coverprofile=cover.out

run:
    @go run cmd/shutter/main.go

clean:
    @rm -rf ./__snapshots__

tui:
    @pushd ./cmd/tui && go build -o shutter ./main.go && popd
    @./cmd/tui/shutter

review:
    @./cmd/tui/shutter
