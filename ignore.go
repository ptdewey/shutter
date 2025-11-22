package shutter

import (
	"regexp"
	"slices"
	"strings"
)

// exactKeyValueIgnore ignores exact key-value matches.
type exactKeyValueIgnore struct {
	key   string
	value string
}

func (e *exactKeyValueIgnore) isOption() {}

func (e *exactKeyValueIgnore) ShouldIgnore(key, value string) bool {
	return e.key == key && (e.value == "*" || e.value == value)
}

// IgnoreKeyValue creates an ignore pattern that matches exact key-value pairs.
// Use "*" as the value to ignore any value for the given key.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreKeyValue("password", "*"),
//	    shutter.IgnoreKeyValue("status", "pending"),
//	)
func IgnoreKeyValue(key, value string) IgnorePattern {
	return &exactKeyValueIgnore{
		key:   key,
		value: value,
	}
}

// regexKeyValueIgnore ignores key-value pairs matching regex patterns.
type regexKeyValueIgnore struct {
	keyPattern   *regexp.Regexp
	valuePattern *regexp.Regexp
}

func (r *regexKeyValueIgnore) isOption() {}

func (r *regexKeyValueIgnore) ShouldIgnore(key, value string) bool {
	keyMatch := r.keyPattern == nil || r.keyPattern.MatchString(key)
	valueMatch := r.valuePattern == nil || r.valuePattern.MatchString(value)
	return keyMatch && valueMatch
}

// IgnoreKeyPattern creates an ignore pattern using regex patterns for keys and values.
// Pass empty string for keyPattern or valuePattern to match any key or value.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreKeyPattern(`.*password.*`, ""),
//	    shutter.IgnoreKeyPattern(`.*token.*`, ""),
//	)
func IgnoreKeyPattern(keyPattern, valuePattern string) IgnorePattern {
	var keyRe, valueRe *regexp.Regexp
	if keyPattern != "" {
		keyRe = regexp.MustCompile(keyPattern)
	}
	if valuePattern != "" {
		valueRe = regexp.MustCompile(valuePattern)
	}
	return &regexKeyValueIgnore{
		keyPattern:   keyRe,
		valuePattern: valueRe,
	}
}

// keyOnlyIgnore ignores any key matching the pattern, regardless of value.
type keyOnlyIgnore struct {
	keys []string
}

func (k *keyOnlyIgnore) isOption() {}

func (k *keyOnlyIgnore) ShouldIgnore(key, value string) bool {
	return slices.Contains(k.keys, key)
}

// IgnoreKey creates an ignore pattern that ignores the specified keys
// regardless of their values.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreKey("password", "secret", "token"),
//	)
func IgnoreKey(keys ...string) IgnorePattern {
	return &keyOnlyIgnore{
		keys: keys,
	}
}

// regexKeyIgnore ignores keys matching a regex pattern.
type regexKeyIgnore struct {
	pattern *regexp.Regexp
}

func (r *regexKeyIgnore) isOption() {}

func (r *regexKeyIgnore) ShouldIgnore(key, value string) bool {
	return r.pattern.MatchString(key)
}

// IgnoreKeyMatching creates an ignore pattern that ignores keys matching
// the given regex pattern.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreKeyMatching(`^user_`),
//	)
func IgnoreKeyMatching(pattern string) IgnorePattern {
	re := regexp.MustCompile(pattern)
	return &regexKeyIgnore{
		pattern: re,
	}
}

// Common ignore patterns for sensitive data
var sensitiveKeys = []string{
	"password", "secret", "token", "api_key", "apiKey",
	"access_token", "refresh_token", "private_key", "privateKey",
	"authorization", "auth", "credentials", "passwd",
}

// IgnoreSensitive ignores common sensitive key names like password, token, etc.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreSensitive(),
//	)
func IgnoreSensitive() IgnorePattern {
	return &keyOnlyIgnore{
		keys: sensitiveKeys,
	}
}

// valueOnlyIgnore ignores any value matching the pattern, regardless of key.
type valueOnlyIgnore struct {
	values []string
}

func (v *valueOnlyIgnore) isOption() {}

func (v *valueOnlyIgnore) ShouldIgnore(key, value string) bool {
	return slices.Contains(v.values, value)
}

// IgnoreValue creates an ignore pattern that ignores the specified values
// regardless of their keys.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreValue("pending", "processing"),
//	)
func IgnoreValue(values ...string) IgnorePattern {
	return &valueOnlyIgnore{
		values: values,
	}
}

// customIgnore allows users to provide a custom ignore function.
type customIgnore struct {
	ignoreFunc func(key, value string) bool
}

func (c *customIgnore) isOption() {}

func (c *customIgnore) ShouldIgnore(key, value string) bool {
	return c.ignoreFunc(key, value)
}

// IgnoreWith creates an ignore pattern using a custom function.
// The function receives the key and value and should return true if the
// key-value pair should be ignored.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreWith(func(key, value string) bool {
//	        return strings.HasPrefix(key, "temp_")
//	    }),
//	)
func IgnoreWith(ignoreFunc func(key, value string) bool) IgnorePattern {
	return &customIgnore{
		ignoreFunc: ignoreFunc,
	}
}

// IgnoreEmpty ignores fields with empty string values.
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreEmpty(),
//	)
func IgnoreEmpty() IgnorePattern {
	return IgnoreWith(func(key, value string) bool {
		return strings.TrimSpace(value) == ""
	})
}

// IgnoreNull ignores fields with null/nil values (represented as "null" in JSON).
//
// This option only works with SnapJSON.
//
// Example:
//
//	shutter.SnapJSON(t, "response", jsonStr,
//	    shutter.IgnoreNull(),
//	)
func IgnoreNull() IgnorePattern {
	return IgnoreWith(func(key, value string) bool {
		return value == "null" || value == "<nil>"
	})
}
