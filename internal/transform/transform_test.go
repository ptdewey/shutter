package transform

import (
	"regexp"
	"strings"
	"testing"
)

// Mock implementations for testing

type mockScrubber struct {
	fn func(string) string
}

func (m *mockScrubber) Scrub(content string) string {
	return m.fn(content)
}

type mockIgnorePattern struct {
	fn func(string, string) bool
}

func (m *mockIgnorePattern) ShouldIgnore(key, value string) bool {
	return m.fn(key, value)
}

// Tests for ApplyScrubbers

func TestApplyScrubbers_NoScrubbers(t *testing.T) {
	input := "hello world"
	result := ApplyScrubbers(input, nil)

	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

func TestApplyScrubbers_SingleScrubber(t *testing.T) {
	scrubber := &mockScrubber{
		fn: func(s string) string {
			return strings.ReplaceAll(s, "secret", "<REDACTED>")
		},
	}

	input := "my secret password"
	expected := "my <REDACTED> password"
	result := ApplyScrubbers(input, []Scrubber{scrubber})

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyScrubbers_MultipleScrubbers(t *testing.T) {
	scrubber1 := &mockScrubber{
		fn: func(s string) string {
			return strings.ReplaceAll(s, "foo", "FOO")
		},
	}
	scrubber2 := &mockScrubber{
		fn: func(s string) string {
			return strings.ReplaceAll(s, "bar", "BAR")
		},
	}

	input := "foo and bar"
	expected := "FOO and BAR"
	result := ApplyScrubbers(input, []Scrubber{scrubber1, scrubber2})

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyScrubbers_OrderMatters(t *testing.T) {
	// Test that scrubbers are applied in order
	scrubber1 := &mockScrubber{
		fn: func(s string) string {
			return strings.ReplaceAll(s, "test", "TEST")
		},
	}
	scrubber2 := &mockScrubber{
		fn: func(s string) string {
			return strings.ReplaceAll(s, "TEST", "FINAL")
		},
	}

	input := "test value"
	expected := "FINAL value"
	result := ApplyScrubbers(input, []Scrubber{scrubber1, scrubber2})

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// Tests for TransformJSON

func TestTransformJSON_InvalidJSON(t *testing.T) {
	config := &Config{}
	_, err := TransformJSON("not valid json", config)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to unmarshal JSON") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTransformJSON_EmptyConfig(t *testing.T) {
	config := &Config{}
	input := `{"name":"John","age":30}`

	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return pretty-printed JSON
	expected := "{\n  \"age\": 30,\n  \"name\": \"John\"\n}"
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestTransformJSON_WithScrubbers(t *testing.T) {
	scrubber := &mockScrubber{
		fn: func(s string) string {
			return strings.ReplaceAll(s, "John", "<NAME>")
		},
	}

	config := &Config{
		Scrubbers: []Scrubber{scrubber},
	}

	input := `{"name":"John","age":30}`
	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<NAME>") {
		t.Errorf("expected scrubber to be applied, got: %s", result)
	}
	if strings.Contains(result, "John") {
		t.Errorf("scrubber failed to replace 'John', got: %s", result)
	}
}

func TestTransformJSON_WithIgnorePatterns(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "password"
		},
	}

	config := &Config{
		Ignore: []IgnorePattern{ignorePattern},
	}

	input := `{"username":"john","password":"secret"}`
	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "password") {
		t.Errorf("expected 'password' to be ignored, got: %s", result)
	}
	if !strings.Contains(result, "username") {
		t.Errorf("expected 'username' to be present, got: %s", result)
	}
}

func TestTransformJSON_ComplexNested(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "secret"
		},
	}

	config := &Config{
		Ignore: []IgnorePattern{ignorePattern},
	}

	input := `{
		"user": {
			"name": "John",
			"secret": "hidden",
			"nested": {
				"secret": "also_hidden",
				"public": "visible"
			}
		}
	}`

	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "hidden") || strings.Contains(result, "also_hidden") {
		t.Errorf("expected nested secrets to be ignored, got: %s", result)
	}
	if !strings.Contains(result, "visible") {
		t.Errorf("expected 'visible' to be present, got: %s", result)
	}
}

func TestTransformJSON_WithArrays(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "id"
		},
	}

	config := &Config{
		Ignore: []IgnorePattern{ignorePattern},
	}

	input := `{
		"users": [
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"}
		]
	}`

	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should remove id fields from all array elements
	if strings.Contains(result, "\"id\"") {
		t.Errorf("expected 'id' fields to be ignored in arrays, got: %s", result)
	}
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "Bob") {
		t.Errorf("expected names to be present, got: %s", result)
	}
}

func TestTransformJSON_ScrubbersAndIgnoreCombined(t *testing.T) {
	scrubber := &mockScrubber{
		fn: func(s string) string {
			re := regexp.MustCompile(`\d{3}-\d{3}-\d{4}`)
			return re.ReplaceAllString(s, "<PHONE>")
		},
	}

	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "ssn"
		},
	}

	config := &Config{
		Scrubbers: []Scrubber{scrubber},
		Ignore:    []IgnorePattern{ignorePattern},
	}

	input := `{"name":"John","phone":"555-123-4567","ssn":"123-45-6789"}`
	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SSN field should be completely removed
	if strings.Contains(result, "ssn") {
		t.Errorf("expected 'ssn' field to be ignored, got: %s", result)
	}

	// Phone should be scrubbed
	if !strings.Contains(result, "<PHONE>") {
		t.Errorf("expected phone to be scrubbed, got: %s", result)
	}
	if strings.Contains(result, "555-123-4567") {
		t.Errorf("expected phone number to be replaced, got: %s", result)
	}
}

// Tests for helper functions

func TestValueToString_String(t *testing.T) {
	result := valueToString("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestValueToString_Nil(t *testing.T) {
	result := valueToString(nil)
	if result != "null" {
		t.Errorf("expected 'null', got %q", result)
	}
}

func TestValueToString_BoolTrue(t *testing.T) {
	result := valueToString(true)
	if result != "true" {
		t.Errorf("expected 'true', got %q", result)
	}
}

func TestValueToString_BoolFalse(t *testing.T) {
	result := valueToString(false)
	if result != "false" {
		t.Errorf("expected 'false', got %q", result)
	}
}

func TestValueToString_Float64(t *testing.T) {
	result := valueToString(42.5)
	if result != "42.5" {
		t.Errorf("expected '42.5', got %q", result)
	}
}

func TestValueToString_Int(t *testing.T) {
	result := valueToString(42)
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
	}
}

func TestValueToString_ComplexType(t *testing.T) {
	// Map should be marshalled to JSON
	m := map[string]any{"key": "value"}
	result := valueToString(m)
	if result != `{"key":"value"}` {
		t.Errorf("expected JSON string, got %q", result)
	}
}

func TestFilterMap_RemovesMatchingKeys(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "remove_me"
		},
	}

	input := map[string]any{
		"keep":      "value1",
		"remove_me": "value2",
		"also_keep": "value3",
	}

	result := filterMap(input, []IgnorePattern{ignorePattern})

	if _, exists := result["remove_me"]; exists {
		t.Error("expected 'remove_me' to be filtered out")
	}
	if result["keep"] != "value1" {
		t.Error("expected 'keep' to remain")
	}
	if result["also_keep"] != "value3" {
		t.Error("expected 'also_keep' to remain")
	}
}

func TestFilterMap_NestedStructures(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "secret"
		},
	}

	input := map[string]any{
		"public": "data",
		"nested": map[string]any{
			"secret": "hidden",
			"public": "visible",
		},
	}

	result := filterMap(input, []IgnorePattern{ignorePattern})

	nested, ok := result["nested"].(map[string]any)
	if !ok {
		t.Fatal("expected nested map")
	}

	if _, exists := nested["secret"]; exists {
		t.Error("expected nested 'secret' to be filtered out")
	}
	if nested["public"] != "visible" {
		t.Error("expected nested 'public' to remain")
	}
}

func TestFilterSlice_ProcessesAllElements(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return key == "id"
		},
	}

	input := []any{
		map[string]any{"id": "1", "name": "Alice"},
		map[string]any{"id": "2", "name": "Bob"},
	}

	result := filterSlice(input, []IgnorePattern{ignorePattern})

	if len(result) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result))
	}

	for i, item := range result {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected map at index %d", i)
		}
		if _, exists := m["id"]; exists {
			t.Errorf("expected 'id' to be filtered at index %d", i)
		}
		if m["name"] == "" {
			t.Errorf("expected 'name' to remain at index %d", i)
		}
	}
}

func TestWalkAndFilter_HandlesAllTypes(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return false // Don't ignore anything
		},
	}

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "string passthrough",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "number passthrough",
			input:    42,
			expected: 42,
		},
		{
			name:     "bool passthrough",
			input:    true,
			expected: true,
		},
		{
			name:     "nil passthrough",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := walkAndFilter(tt.input, []IgnorePattern{ignorePattern})
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTransformJSON_EmptyObject(t *testing.T) {
	config := &Config{}
	input := `{}`

	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "{}"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTransformJSON_EmptyArray(t *testing.T) {
	config := &Config{}
	input := `[]`

	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[]"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTransformJSON_IgnoreByValue(t *testing.T) {
	ignorePattern := &mockIgnorePattern{
		fn: func(key, value string) bool {
			return value == "ignore_this"
		},
	}

	config := &Config{
		Ignore: []IgnorePattern{ignorePattern},
	}

	input := `{"field1":"keep","field2":"ignore_this","field3":"also_keep"}`
	result, err := TransformJSON(input, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(result, "field2") {
		t.Errorf("expected field with ignored value to be removed, got: %s", result)
	}
	if !strings.Contains(result, "field1") || !strings.Contains(result, "field3") {
		t.Errorf("expected other fields to remain, got: %s", result)
	}
}
