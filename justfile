clean-test:
    @rm -rf ./__snapshots__
    @go test ./... -cover -coverprofile=cover.out

test:
    @go test ./... -cover -coverprofile=cover.out


clean:
    @rm -rf ./__snapshots__
