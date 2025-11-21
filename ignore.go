package freeze

import (
	"regexp"
	"slices"
	"strings"
)

// IgnorePattern determines whether a key-value pair should be excluded
// from the snapshot. This is primarily used for JSON and map structures.
type IgnorePattern interface {
	ShouldIgnore(key, value string) bool
}

// exactKeyValueIgnore ignores exact key-value matches.
type exactKeyValueIgnore struct {
	key   string
	value string
}

func (e *exactKeyValueIgnore) ShouldIgnore(key, value string) bool {
	return e.key == key && (e.value == "*" || e.value == value)
}

// IgnoreKeyValue creates an ignore pattern that matches exact key-value pairs.
// Use "*" as the value to ignore any value for the given key.
func IgnoreKeyValue(key, value string) SnapshotOption {
	return WithIgnorePattern(&exactKeyValueIgnore{
		key:   key,
		value: value,
	})
}

// regexKeyValueIgnore ignores key-value pairs matching regex patterns.
type regexKeyValueIgnore struct {
	keyPattern   *regexp.Regexp
	valuePattern *regexp.Regexp
}

func (r *regexKeyValueIgnore) ShouldIgnore(key, value string) bool {
	keyMatch := r.keyPattern == nil || r.keyPattern.MatchString(key)
	valueMatch := r.valuePattern == nil || r.valuePattern.MatchString(value)
	return keyMatch && valueMatch
}

// IgnoreKeyPattern creates an ignore pattern using regex patterns for keys and values.
// Pass empty string for keyPattern or valuePattern to match any key or value.
func IgnoreKeyPattern(keyPattern, valuePattern string) SnapshotOption {
	var keyRe, valueRe *regexp.Regexp
	if keyPattern != "" {
		keyRe = regexp.MustCompile(keyPattern)
	}
	if valuePattern != "" {
		valueRe = regexp.MustCompile(valuePattern)
	}
	return WithIgnorePattern(&regexKeyValueIgnore{
		keyPattern:   keyRe,
		valuePattern: valueRe,
	})
}

// keyOnlyIgnore ignores any key matching the pattern, regardless of value.
type keyOnlyIgnore struct {
	keys []string
}

func (k *keyOnlyIgnore) ShouldIgnore(key, value string) bool {
	return slices.Contains(k.keys, key)
}

// IgnoreKeys creates an ignore pattern that ignores the specified keys
// regardless of their values.
func IgnoreKeys(keys ...string) SnapshotOption {
	return WithIgnorePattern(&keyOnlyIgnore{
		keys: keys,
	})
}

// regexKeyIgnore ignores keys matching a regex pattern.
type regexKeyIgnore struct {
	pattern *regexp.Regexp
}

func (r *regexKeyIgnore) ShouldIgnore(key, value string) bool {
	return r.pattern.MatchString(key)
}

// IgnoreKeysMatching creates an ignore pattern that ignores keys matching
// the given regex pattern.
func IgnoreKeysMatching(pattern string) SnapshotOption {
	re := regexp.MustCompile(pattern)
	return WithIgnorePattern(&regexKeyIgnore{
		pattern: re,
	})
}

// Common ignore patterns for sensitive data
var sensitiveKeys = []string{
	"password", "secret", "token", "api_key", "apiKey",
	"access_token", "refresh_token", "private_key", "privateKey",
	"authorization", "auth", "credentials", "passwd",
}

// IgnoreSensitiveKeys ignores common sensitive key names like password, token, etc.
func IgnoreSensitiveKeys() SnapshotOption {
	return WithIgnorePattern(&keyOnlyIgnore{
		keys: sensitiveKeys,
	})
}

// valueOnlyIgnore ignores any value matching the pattern, regardless of key.
type valueOnlyIgnore struct {
	values []string
}

func (v *valueOnlyIgnore) ShouldIgnore(key, value string) bool {
	return slices.Contains(v.values, value)
}

// IgnoreValues creates an ignore pattern that ignores the specified values
// regardless of their keys.
func IgnoreValues(values ...string) SnapshotOption {
	return WithIgnorePattern(&valueOnlyIgnore{
		values: values,
	})
}

// customIgnore allows users to provide a custom ignore function.
type customIgnore struct {
	ignoreFunc func(key, value string) bool
}

func (c *customIgnore) ShouldIgnore(key, value string) bool {
	return c.ignoreFunc(key, value)
}

// CustomIgnore creates an ignore pattern using a custom function.
func CustomIgnore(ignoreFunc func(key, value string) bool) SnapshotOption {
	return WithIgnorePattern(&customIgnore{
		ignoreFunc: ignoreFunc,
	})
}

// IgnoreEmptyValues ignores fields with empty string values.
func IgnoreEmptyValues() SnapshotOption {
	return CustomIgnore(func(key, value string) bool {
		return strings.TrimSpace(value) == ""
	})
}

// IgnoreNullValues ignores fields with null/nil values (represented as "null" in JSON).
func IgnoreNullValues() SnapshotOption {
	return CustomIgnore(func(key, value string) bool {
		return value == "null" || value == "<nil>"
	})
}
