# Freeze

A [birdie](https://github.com/giacomocavalieri/birdie) and [insta](https://github.com/mitsuhiko/insta) inspired snapshot testing library for Go.

![New snapshot screen](./assets/screenshot-new.png "New snapshot view")

![Snapshot review CLI](./assets/screenshots-diff-cli.png "Snapshot diff view (CLI)")

## Installation

```sh
go get github.com/ptdewey/freeze
```

## Usage

```go
package yourpackage_test

func TestSomething(t *testing.T) {
    result := SomeFunction("foo")
    freeze.Snap(t, result)

    // To capture the calling function name use SnapFunc
    freeze.SnapFunc(t, SomeFunction("bar"))
}
```

To review a set of snapshots, run:

```sh
go run github.com/ptdewey/freeze/cmd/freeze review
```

<!-- TODO: add example of `freeze.Review()` in go code -->

Freeze also includes (in a separate Go module) a [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI in [cmd/tui/main.go](./cmd/tui/main.go). (The TUI is shipped in a separate module to make the added dependencies optional)

### TUI Usage

```sh
go run github.com/ptdewey/freeze/cmd/tui review
```

#### Interactive Controls

- `a` - Accept current snapshot
- `r` - Reject current snapshot
- `s` - Skip current snapshot
- `A` - Accept all remaining snapshots
- `R` - Reject all remaining snapshots
- `S` - Skip all remaining snapshots
- `q` - Quit

#### Alternative Commands

```sh
# Accept all new snapshots without review
go run github.com/ptdewey/freeze/cmd/tui accept-all

# Reject all new snapshots without review
go run github.com/ptdewey/freeze/cmd/tui reject-all
```

## Disclaimer

- This package was largely vibe coded, your mileage may vary (but this library provides more of what I want than the ones below).

## Other Libraries

- [go-snaps](https://github.com/gkampitakis/go-snaps)
  - Freeze uses the diff implementation from `go-snaps`.
- [cupaloy](https://github.com/bradleyjkemp/cupaloy)
