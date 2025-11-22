package shutter_test

import (
	"testing"

	"github.com/ptdewey/shutter"
)

func TestIgnoreKeyValue(t *testing.T) {
	jsonStr := `{
		"username": "john_doe",
		"password": "secret123",
		"email": "john@example.com",
		"api_key": "sk_live_abc123"
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Password Field", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreKeyValue("password", "*"),
		shutter.IgnoreKeyValue("api_key", "*"),
	})
}

func TestIgnoreKeys(t *testing.T) {
	jsonStr := `{
		"id": 1,
		"name": "John Doe",
		"password": "secret",
		"secret": "confidential",
		"token": "abc123",
		"email": "john@example.com"
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Multiple Keys", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreKeys("password", "secret", "token"),
	})
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

	shutter.SnapJSONWithOptions(t, "Ignore Sensitive Keys", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreSensitiveKeys(),
	})
}

func TestIgnoreKeysMatching(t *testing.T) {
	jsonStr := `{
		"user_id": 1,
		"user_name": "john",
		"user_email": "john@example.com",
		"product_id": 100,
		"product_name": "Widget"
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Keys Matching Pattern", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreKeysMatching(`^user_`),
	})
}

func TestIgnoreKeyPattern(t *testing.T) {
	jsonStr := `{
		"username": "john_doe",
		"password": "secret",
		"admin_password": "admin_secret",
		"user_token": "token123",
		"email": "john@example.com"
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Key Pattern", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreKeyPattern(`.*password.*`, ""),
		shutter.IgnoreKeyPattern(`.*token.*`, ""),
	})
}

func TestIgnoreValues(t *testing.T) {
	jsonStr := `{
		"status": "pending",
		"result": "pending",
		"message": "Processing",
		"state": "pending"
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Specific Values", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreValues("pending"),
	})
}

func TestIgnoreEmptyValues(t *testing.T) {
	jsonStr := `{
		"name": "John Doe",
		"middle_name": "",
		"nickname": "   ",
		"email": "john@example.com",
		"phone": ""
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Empty Values", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreEmptyValues(),
	})
}

func TestIgnoreNullValues(t *testing.T) {
	jsonStr := `{
		"name": "John Doe",
		"middle_name": null,
		"email": "john@example.com",
		"phone": null,
		"age": 30
	}`

	shutter.SnapJSONWithOptions(t, "Ignore Null Values", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreNullValues(),
	})
}

func TestCustomIgnore(t *testing.T) {
	jsonStr := `{
		"id": 1,
		"name": "John Doe",
		"age": 25,
		"score": 95,
		"grade": "A"
	}`

	shutter.SnapJSONWithOptions(t, "Custom Ignore Function", jsonStr, []shutter.SnapshotOption{
		shutter.CustomIgnore(func(key, value string) bool {
			// Ignore numeric values
			return value == "1" || value == "25" || value == "95"
		}),
	})
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

	shutter.SnapJSONWithOptions(t, "Nested Ignore Patterns", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreSensitiveKeys(),
	})
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

	shutter.SnapJSONWithOptions(t, "Combined Ignore and Scrub", jsonStr, []shutter.SnapshotOption{
		// Ignore sensitive keys entirely
		shutter.IgnoreKeys("password", "api_key"),
		// Scrub dynamic/identifiable data
		shutter.ScrubUUIDs(),
		shutter.ScrubEmails(),
		shutter.ScrubTimestamps(),
		shutter.ScrubIPAddresses(),
	})
}

func TestIgnoreInArrays(t *testing.T) {
	jsonStr := `{
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
	}`

	shutter.SnapJSONWithOptions(t, "Ignore in Arrays", jsonStr, []shutter.SnapshotOption{
		shutter.IgnoreKeys("password"),
	})
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

	shutter.SnapJSONWithOptions(t, "Real World API Response", jsonStr, []shutter.SnapshotOption{
		// Ignore sensitive fields
		shutter.IgnoreSensitiveKeys(),
		shutter.IgnoreKeys("card_number"),
		// Scrub dynamic/identifiable data
		shutter.ScrubUUIDs(),
		shutter.ScrubEmails(),
		shutter.ScrubTimestamps(),
		shutter.ScrubIPAddresses(),
		shutter.ScrubJWTs(),
	})
}
