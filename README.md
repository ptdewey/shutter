# Shutter

A [birdie](https://github.com/giacomocavalieri/birdie) and [insta](https://github.com/mitsuhiko/insta) inspired snapshot testing library for Go.

![New snapshot screen](./assets/screenshot-new.png "New snapshot view")

![Snapshot review CLI](./assets/screenshot-diff-cli.png "Snapshot diff view (CLI)")

## Installation

```sh
go get github.com/ptdewey/shutter
```

## Usage

### Basic Usage

```go
package package_test

func TestSomething(t *testing.T) {
    result := SomeFunction("foo")
    shutter.Snap(t, "test title", result)
}
```

### Snapshotting Multiple Values

Use `SnapMany()` when you need to snapshot multiple related values together:

```go
func TestMultipleValues(t *testing.T) {
    request := buildRequest()
    response := handleRequest(request)

    // Snapshot both request and response together
    shutter.SnapMany(t, "title", []any{request, response})
}
```

### Advanced Usage: Scrubbers and Ignore Patterns

shutter supports data scrubbing and field filtering to handle dynamic or sensitive data in snapshots.

#### Scrubbers

Scrubbers transform content before snapshotting, typically to replace dynamic or sensitive data with placeholders:

```go
func TestUserAPI(t *testing.T) {
    user := api.GetUser("123")

    // Replace UUIDs and timestamps with placeholders
    shutter.Snap(t, "user", user,
        shutter.ScrubUUID(),
        shutter.ScrubTimestamp(),
    )
}
```

**Built-in Scrubbers:**

- `ScrubUUID()` - Replaces UUIDs with `<UUID>`
- `ScrubTimestamp()` - Replaces ISO8601 timestamps with `<TIMESTAMP>`
- `ScrubEmail()` - Replaces email addresses with `<EMAIL>`
- `ScrubIP()` - Replaces IPv4 addresses with `<IP>`
- `ScrubJWT()` - Replaces JWT tokens with `<JWT>`
- `ScrubCreditCard()` - Replaces credit card numbers with `<CREDIT_CARD>`
- `ScrubAPIKey()` - Replaces API keys with `<API_KEY>`
- `ScrubDate()` - Replaces various date formats with `<DATE>`
- `ScrubUnixTimestamp()` - Replaces Unix timestamps with `<UNIX_TS>`

**Custom Scrubbers:**

```go
// Using regex patterns
shutter.ScrubRegex(`user-\d+`, "<USER_ID>")

// Using exact string matching
shutter.ScrubExact("secret_value", "<REDACTED>")

// Using custom functions
shutter.ScrubWith(func(content string) string {
    return strings.ReplaceAll(content, "localhost", "<HOST>")
})
```

#### Ignore Patterns

Ignore patterns remove specific fields from JSON structures before snapshotting:

```go
func TestAPIResponse(t *testing.T) {
    response := api.GetData()
    jsonBytes, _ := json.Marshal(response)

    // Ignore sensitive fields and null values
    shutter.SnapJSON(t, "response", string(jsonBytes),
        shutter.IgnoreSensitive(),
        shutter.IgnoreNull(),
        shutter.IgnoreKey("created_at", "updated_at"),
    )
}
```

**Built-in Ignore Patterns:**

- `IgnoreSensitive()` - Ignores common sensitive keys (password, token, api_key, etc.)
- `IgnoreEmpty()` - Ignores fields with empty string values
- `IgnoreNull()` - Ignores fields with null values

**Custom Ignore Patterns:**

```go
// Ignore specific keys
shutter.IgnoreKey("id", "timestamp", "version")

// Ignore key-value pairs
shutter.IgnoreKeyValue("status", "pending")

// Ignore keys matching a regex pattern
shutter.IgnoreKeyMatching(`^_.*`) // Ignore all keys starting with underscore

// Ignore specific values
shutter.IgnoreValue("null", "undefined", "")

// Using custom functions
shutter.IgnoreWith(func(key, value string) bool {
    return strings.HasPrefix(key, "temp_")
})
```

#### Combining Options

You can combine multiple scrubbers and ignore patterns:

```go
func TestComplexData(t *testing.T) {
    data := generateTestData()
    jsonBytes, _ := json.Marshal(data)

    shutter.SnapJSON(t, "data", string(jsonBytes),
        // First, remove unwanted fields
        shutter.IgnoreSensitive(),
        shutter.IgnoreKey("debug_info"),
        shutter.IgnoreNull(),

        // Then, scrub dynamic values in remaining fields
        shutter.ScrubUUID(),
        shutter.ScrubTimestamp(),
        shutter.ScrubEmail(),
    )
}
```

**Note:** Ignore patterns only work with `SnapJSON()`. Use scrubbers with `Snap()`, `SnapMany()`, or `SnapString()`.

#### API Reference

**Snapshot Functions:**

```go
// For single values (structs, maps, slices, etc.)
shutter.Snap(t, "title", value, options...)

// For multiple related values
shutter.SnapMany(t, "title", []any{value1, value2, value3}, options...)

// For JSON strings (supports both scrubbers and ignore patterns)
shutter.SnapJSON(t, "title", jsonString, options...)

// For plain strings
shutter.SnapString(t, "title", content, options...)
```

### Reviewing Snapshots

To review a set of snapshots, run:

```sh
go run github.com/ptdewey/shutter/cmd/shutter review
```

Shutter can also be used programmatically:

```go
// Example: tools/shutter/main.go
package main

import "github.com/ptdewey/shutter"

func main() {
    // This will start the CLI review tool
    shutter.Review()
}
```

Which can then be run with:

```sh
go run tools/shutter/main.go
```

Shutter also includes (in a separate Go module) a [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI in [cmd/tui/main.go](./cmd/tui/main.go).
(The TUI is shipped in a separate module to make the added dependencies optional)

### TUI Usage

```sh
go run github.com/ptdewey/shutter/cmd/tui review
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
go run github.com/ptdewey/shutter/cmd/tui accept-all

# Reject all new snapshots without review
go run github.com/ptdewey/shutter/cmd/tui reject-all
```

## Other Libraries

- [go-snaps](https://github.com/gkampitakis/go-snaps)
  - shutter uses the diff implementation from `go-snaps`.
- [cupaloy](https://github.com/bradleyjkemp/cupaloy)
