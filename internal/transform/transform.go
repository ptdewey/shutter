package transform

import (
	"encoding/json"
	"fmt"
)

// Config holds the transformation configuration.
type Config struct {
	Scrubbers []Scrubber
	Ignore    []IgnorePattern
}

// Scrubber transforms content before snapshotting.
type Scrubber interface {
	Scrub(content string) string
}

// IgnorePattern determines whether a key-value pair should be excluded.
type IgnorePattern interface {
	ShouldIgnore(key, value string) bool
}

// ApplyScrubbers applies all scrubbers to the content in order.
func ApplyScrubbers(content string, scrubbers []Scrubber) string {
	result := content
	for _, scrubber := range scrubbers {
		result = scrubber.Scrub(result)
	}
	return result
}

// TransformJSON applies scrubbers and ignore patterns to JSON data.
func TransformJSON(jsonStr string, config *Config) (string, error) {
	if config == nil || (len(config.Scrubbers) == 0 && len(config.Ignore) == 0) {
		return jsonStr, nil
	}

	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Apply ignore patterns first (removes fields)
	if len(config.Ignore) > 0 {
		data = walkAndFilter(data, config.Ignore)
	}

	// Marshal back to JSON
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	result := string(prettyJSON)

	// Apply scrubbers to the final string
	result = ApplyScrubbers(result, config.Scrubbers)

	return result, nil
}

// walkAndFilter recursively walks the data structure and filters out ignored fields.
func walkAndFilter(data any, ignorePatterns []IgnorePattern) any {
	switch v := data.(type) {
	case map[string]any:
		return filterMap(v, ignorePatterns)
	case []any:
		return filterSlice(v, ignorePatterns)
	default:
		return data
	}
}

// filterMap filters a map, removing entries that match ignore patterns.
func filterMap(m map[string]any, ignorePatterns []IgnorePattern) map[string]any {
	result := make(map[string]any)
	for key, value := range m {
		// Convert value to string for comparison
		valueStr := valueToString(value)

		// Check if this key-value pair should be ignored
		shouldIgnore := false
		for _, pattern := range ignorePatterns {
			if pattern.ShouldIgnore(key, valueStr) {
				shouldIgnore = true
				break
			}
		}

		if !shouldIgnore {
			// Recursively filter nested structures
			result[key] = walkAndFilter(value, ignorePatterns)
		}
	}
	return result
}

// filterSlice filters a slice, recursively processing each element.
func filterSlice(s []any, ignorePatterns []IgnorePattern) []any {
	result := make([]any, len(s))
	for i, item := range s {
		result[i] = walkAndFilter(item, ignorePatterns)
	}
	return result
}

// valueToString converts various value types to string for comparison.
func valueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", v)
	case int, int64:
		return fmt.Sprintf("%d", v)
	default:
		// For complex types, marshal to JSON
		if bytes, err := json.Marshal(v); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v)
	}
}
