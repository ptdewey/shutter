# Freeze

A [birdie](https://github.com/giacomocavalieri/birdie) and [insta](https://github.com/mitsuhiko/insta) inspired snapshot testing library for Go.

![New snapshot screen](./assets/screenshot-new.png "New snapshot view")

![Snapshot review CLI](./assets/screenshot-diff-cli.png "Snapshot diff view (CLI)")

## Installation

```sh
go get github.com/ptdewey/freeze
```

## Usage

### Basic Usage

```go
package package_test

func TestSomething(t *testing.T) {
    result := SomeFunction("foo")
    freeze.Snap(t, result)
}
```

### Advanced Usage: Scrubbers and Ignore Patterns

Freeze supports data scrubbing and field filtering to handle dynamic or sensitive data in snapshots.

#### Scrubbers

Scrubbers transform content before snapshotting, typically to replace dynamic or sensitive data with placeholders:

```go
func TestUserAPI(t *testing.T) {
    user := api.GetUser("123")
    
    // Replace UUIDs and timestamps with placeholders
    freeze.SnapWithOptions(t, "user", []freeze.SnapshotOption{
        freeze.ScrubUUIDs(),
        freeze.ScrubTimestamps(),
    }, user)
}
```

**Built-in Scrubbers:**
- `ScrubUUIDs()` - Replaces UUIDs with `<UUID>`
- `ScrubTimestamps()` - Replaces ISO8601 timestamps with `<TIMESTAMP>`
- `ScrubEmails()` - Replaces email addresses with `<EMAIL>`
- `ScrubIPAddresses()` - Replaces IPv4 addresses with `<IP>`
- `ScrubJWTs()` - Replaces JWT tokens with `<JWT>`
- `ScrubCreditCards()` - Replaces credit card numbers with `<CREDIT_CARD>`
- `ScrubAPIKeys()` - Replaces API keys with `<API_KEY>`
- `ScrubDates()` - Replaces various date formats with `<DATE>`
- `ScrubUnixTimestamps()` - Replaces Unix timestamps with `<UNIX_TS>`

**Custom Scrubbers:**

```go
// Using regex patterns
freeze.RegexScrubber(`user-\d+`, "<USER_ID>")

// Using exact string matching
freeze.ExactMatchScrubber("secret_value", "<REDACTED>")

// Using custom functions
freeze.CustomScrubber(func(content string) string {
    return strings.ReplaceAll(content, "localhost", "<HOST>")
})
```

#### Ignore Patterns

Ignore patterns remove specific fields from JSON structures before snapshotting:

```go
func TestAPIResponse(t *testing.T) {
    response := api.GetData()
    
    // Ignore sensitive fields and null values
    freeze.SnapJSONWithOptions(t, "response", response, []freeze.SnapshotOption{
        freeze.IgnoreSensitiveKeys(),
        freeze.IgnoreNullValues(),
        freeze.IgnoreKeys("created_at", "updated_at"),
    })
}
```

**Built-in Ignore Patterns:**
- `IgnoreSensitiveKeys()` - Ignores common sensitive keys (password, token, api_key, etc.)
- `IgnoreEmptyValues()` - Ignores fields with empty string values
- `IgnoreNullValues()` - Ignores fields with null values

**Custom Ignore Patterns:**

```go
// Ignore specific keys
freeze.IgnoreKeys("id", "timestamp", "version")

// Ignore key-value pairs
freeze.IgnoreKeyValue("status", "pending")

// Ignore keys matching a regex pattern
freeze.IgnoreKeysMatching(`^_.*`) // Ignore all keys starting with underscore

// Ignore specific values
freeze.IgnoreValues("null", "undefined", "")

// Using custom functions
freeze.CustomIgnore(func(key, value string) bool {
    return strings.HasPrefix(key, "temp_")
})
```

#### Combining Options

You can combine multiple scrubbers and ignore patterns:

```go
func TestComplexData(t *testing.T) {
    data := generateTestData()
    
    freeze.SnapWithOptions(t, "data", []freeze.SnapshotOption{
        // Scrubbers
        freeze.ScrubUUIDs(),
        freeze.ScrubTimestamps(),
        freeze.ScrubEmails(),
        
        // Ignore patterns
        freeze.IgnoreSensitiveKeys(),
        freeze.IgnoreKeys("debug_info"),
        freeze.IgnoreNullValues(),
    }, data)
}
```

#### API Reference

Three snapshot functions support options:

```go
// For general values (structs, maps, slices, etc.)
freeze.SnapWithOptions(t, "title", []freeze.SnapshotOption{...}, value)

// For JSON strings
freeze.SnapJSONWithOptions(t, "title", jsonString, []freeze.SnapshotOption{...})

// For plain strings
freeze.SnapStringWithOptions(t, "title", content, []freeze.SnapshotOption{...})
```

### Reviewing Snapshots

To review a set of snapshots, run:

```sh
go run github.com/ptdewey/freeze/cmd/freeze review
```

Freeze can also be used programmatically:

```go
// Example: tools/freeze/main.go
package main

import "github.com/ptdewey/freeze"

func main() {
    // This will start the CLI review tool
    freeze.Review()
}
```

Which can then be run with:

```sh
go run tools/freeze/main.go
```

Freeze also includes (in a separate Go module) a [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI in [cmd/tui/main.go](./cmd/tui/main.go).
(The TUI is shipped in a separate module to make the added dependencies optional)

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
