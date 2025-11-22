package shutter_test

import (
	"testing"

	"github.com/ptdewey/shutter"
)

func TestIgnoreKeys(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		opts  []shutter.Option
		title string
	}{
		{
			name: "multiple_keys",
			json: `{
				"id": 1,
				"name": "John Doe",
				"password": "secret",
				"secret": "confidential",
				"token": "abc123",
				"email": "john@example.com"
			}`,
			opts:  []shutter.Option{shutter.IgnoreKey("password", "secret", "token")},
			title: "Ignore Multiple Keys",
		},
		{
			name: "key_value_pairs",
			json: `{
				"username": "john_doe",
				"password": "secret123",
				"email": "john@example.com",
				"api_key": "sk_live_abc123"
			}`,
			opts: []shutter.Option{
				shutter.IgnoreKeyValue("password", "*"),
				shutter.IgnoreKeyValue("api_key", "*"),
			},
			title: "Ignore Password Field",
		},
		{
			name: "arrays",
			json: `{
				"users": [
					{
						"id": 1,
						"name": "Alice",
						"password": "secret1",
						"email": "alice@example.com"
					},
					{
						"id": 2,
						"name": "Bob",
						"password": "secret2",
						"email": "bob@example.com"
					}
				]
			}`,
			opts:  []shutter.Option{shutter.IgnoreKey("password")},
			title: "Ignore in Arrays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutter.SnapJSON(t, tt.title, tt.json, tt.opts...)
		})
	}
}

func TestIgnoreSensitiveKeys(t *testing.T) {
	jsonStr := `{
		"username": "john_doe",
		"password": "secret123",
		"api_key": "sk_live_abc123",
		"access_token": "token123",
		"refresh_token": "refresh123",
		"email": "john@example.com",
		"name": "John Doe"
	}`

	shutter.SnapJSON(t, "Ignore Sensitive Keys", jsonStr,
		shutter.IgnoreSensitive(),
	)
}

func TestIgnoreKeyPatterns(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		opts  []shutter.Option
		title string
	}{
		{
			name: "prefix_pattern",
			json: `{
				"user_id": 1,
				"user_name": "john",
				"user_email": "john@example.com",
				"product_id": 100,
				"product_name": "Widget"
			}`,
			opts:  []shutter.Option{shutter.IgnoreKeyMatching(`^user_`)},
			title: "Ignore Keys Matching Pattern",
		},
		{
			name: "contains_pattern",
			json: `{
				"username": "john_doe",
				"password": "secret",
				"admin_password": "admin_secret",
				"user_token": "token123",
				"email": "john@example.com"
			}`,
			opts: []shutter.Option{
				shutter.IgnoreKeyPattern(`.*password.*`, ""),
				shutter.IgnoreKeyPattern(`.*token.*`, ""),
			},
			title: "Ignore Key Pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutter.SnapJSON(t, tt.title, tt.json, tt.opts...)
		})
	}
}

func TestIgnoreValues(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		opts  []shutter.Option
		title string
	}{
		{
			name: "specific_values",
			json: `{
				"status": "pending",
				"result": "pending",
				"message": "Processing",
				"state": "pending"
			}`,
			opts:  []shutter.Option{shutter.IgnoreValue("pending")},
			title: "Ignore Specific Values",
		},
		{
			name: "empty_values",
			json: `{
				"name": "John Doe",
				"middle_name": "",
				"nickname": "   ",
				"email": "john@example.com",
				"phone": ""
			}`,
			opts:  []shutter.Option{shutter.IgnoreEmpty()},
			title: "Ignore Empty Values",
		},
		{
			name: "null_values",
			json: `{
				"name": "John Doe",
				"middle_name": null,
				"email": "john@example.com",
				"phone": null,
				"age": 30
			}`,
			opts:  []shutter.Option{shutter.IgnoreNull()},
			title: "Ignore Null Values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutter.SnapJSON(t, tt.title, tt.json, tt.opts...)
		})
	}
}

func TestCustomIgnore(t *testing.T) {
	jsonStr := `{
		"id": 1,
		"name": "John Doe",
		"age": 25,
		"score": 95,
		"grade": "A"
	}`

	shutter.SnapJSON(t, "Custom Ignore Function", jsonStr,
		shutter.IgnoreWith(func(key, value string) bool {
			// Ignore numeric values
			return value == "1" || value == "25" || value == "95"
		}),
	)
}

func TestNestedIgnorePatterns(t *testing.T) {
	jsonStr := `{
		"user": {
			"id": 1,
			"name": "John Doe",
			"password": "secret",
			"email": "john@example.com",
			"profile": {
				"bio": "Developer",
				"api_key": "sk_live_abc123",
				"website": "https://example.com"
			}
		},
		"admin": {
			"password": "admin_secret",
			"token": "admin_token_123"
		}
	}`

	shutter.SnapJSON(t, "Nested Ignore Patterns", jsonStr,
		shutter.IgnoreSensitive(),
	)
}

func TestCombinedIgnoreAndScrub(t *testing.T) {
	jsonStr := `{
		"user_id": "550e8400-e29b-41d4-a716-446655440000",
		"name": "John Doe",
		"email": "john@example.com",
		"password": "secret123",
		"created_at": "2023-01-15T10:30:00Z",
		"api_key": "sk_live_abc123",
		"ip_address": "192.168.1.1"
	}`

	shutter.SnapJSON(t, "Combined Ignore and Scrub", jsonStr,
		// Ignore sensitive keys entirely
		shutter.IgnoreKey("password", "api_key"),
		// Scrub dynamic/identifiable data
		shutter.ScrubUUID(),
		shutter.ScrubEmail(),
		shutter.ScrubTimestamp(),
		shutter.ScrubIP(),
	)
}

func TestComplexRealWorldExample(t *testing.T) {
	jsonStr := `{
		"request_id": "550e8400-e29b-41d4-a716-446655440000",
		"timestamp": "2023-11-20T15:30:00Z",
		"user": {
			"id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"email": "user@example.com",
			"name": "John Doe",
			"password": "hashed_password",
			"api_key": "sk_live_abc123def456",
			"ip_address": "192.168.1.1",
			"created_at": "2023-01-15T10:30:00Z"
		},
		"transaction": {
			"id": "txn_abc123",
			"amount": 99.99,
			"currency": "USD",
			"card_number": "4532-1234-5678-9010",
			"timestamp": "2023-11-20T15:30:00Z"
		},
		"metadata": {
			"server_ip": "10.0.0.5",
			"session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			"user_agent": "Mozilla/5.0"
		}
	}`

	shutter.SnapJSON(t, "Real World API Response", jsonStr,
		// Ignore sensitive fields
		shutter.IgnoreSensitive(),
		shutter.IgnoreKey("card_number"),
		// Scrub dynamic/identifiable data
		shutter.ScrubUUID(),
		shutter.ScrubEmail(),
		shutter.ScrubTimestamp(),
		shutter.ScrubIP(),
		shutter.ScrubJWT(),
	)
}
