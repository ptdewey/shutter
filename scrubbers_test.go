package shutter_test

import (
	"strings"
	"testing"

	"github.com/ptdewey/shutter"
)

func TestBuiltInScrubbers(t *testing.T) {
	// Test all built-in scrubbers in one comprehensive test
	jsonStr := `{
		"user_id": "550e8400-e29b-41d4-a716-446655440000",
		"session_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"email": "user@example.com",
		"backup_email": "backup.user+tag@subdomain.example.co.uk",
		"created_at": "2023-01-15T10:30:00Z",
		"updated_at": "2023-11-20T15:45:30.123Z",
		"birth_date": "1990-05-15",
		"us_format_date": "12/25/2023",
		"unix_created": 1699999999,
		"unix_updated": 1700000000000,
		"client_ip": "192.168.1.1",
		"server_ip": "10.0.0.5",
		"jwt_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		"stripe_key": "sk_live_51HqZ2bKl4FGBMFpLxO0123",
		"api_key": "api_key_prod_abc123def456",
		"card_number": "4532-1234-5678-9010",
		"backup_card": "4532 1234 5678 9010",
		"name": "John Doe",
		"message": "Connection from 172.16.0.100"
	}`

	shutter.SnapJSON(t, "Multiple Scrubbers", jsonStr,
		shutter.ScrubUUID(),
		shutter.ScrubEmail(),
		shutter.ScrubTimestamp(),
		shutter.ScrubDate(),
		shutter.ScrubUnixTimestamp(),
		shutter.ScrubIP(),
		shutter.ScrubJWT(),
		shutter.ScrubAPIKey(),
		shutter.ScrubCreditCard(),
	)
}

func TestIndividualScrubbers(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		scrubber shutter.Option
		title    string
	}{
		{
			name: "uuid",
			json: `{
				"user_id": "550e8400-e29b-41d4-a716-446655440000",
				"session_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				"name": "John Doe"
			}`,
			scrubber: shutter.ScrubUUID(),
			title:    "Scrubbed UUIDs",
		},
		{
			name: "timestamps",
			json: `{
				"created_at": "2023-01-15T10:30:00Z",
				"updated_at": "2023-11-20T15:45:30.123Z",
				"deleted_at": "2023-12-01T08:00:00+05:00",
				"name": "Test Event"
			}`,
			scrubber: shutter.ScrubTimestamp(),
			title:    "Scrubbed Timestamps",
		},
		{
			name: "emails",
			json: `{
				"email": "user@example.com",
				"backup_email": "backup.user+tag@subdomain.example.co.uk",
				"name": "John Doe"
			}`,
			scrubber: shutter.ScrubEmail(),
			title:    "Scrubbed Emails",
		},
		{
			name: "ip_addresses",
			json: `{
				"client_ip": "192.168.1.1",
				"server_ip": "10.0.0.5",
				"message": "Connection from 172.16.0.100"
			}`,
			scrubber: shutter.ScrubIP(),
			title:    "Scrubbed IPs",
		},
		{
			name: "jwts",
			json: `{
				"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
			}`,
			scrubber: shutter.ScrubJWT(),
			title:    "Scrubbed JWTs",
		},
		{
			name: "dates",
			json: `{
				"birth_date": "1990-05-15",
				"hire_date": "2020-01-01",
				"us_format": "12/25/2023",
				"name": "John Doe"
			}`,
			scrubber: shutter.ScrubDate(),
			title:    "Scrubbed Dates",
		},
		{
			name: "api_keys",
			json: `{
				"stripe_key": "sk_live_51HqZ2bKl4FGBMFpLxO0123",
				"test_key": "pk_test_51HqZ2bKl4FGBMFpLxO0456",
				"api_key_prod": "api_key_prod_abc123def456",
				"name": "Test Config"
			}`,
			scrubber: shutter.ScrubAPIKey(),
			title:    "Scrubbed API Keys",
		},
		{
			name: "credit_cards",
			json: `{
				"card_number": "4532-1234-5678-9010",
				"backup_card": "4532 1234 5678 9010",
				"another_card": "4532123456789010",
				"name": "John Doe"
			}`,
			scrubber: shutter.ScrubCreditCard(),
			title:    "Scrubbed Credit Cards",
		},
		{
			name: "unix_timestamps",
			json: `{
				"created": 1699999999,
				"updated": 1700000000000,
				"deleted": 1700000000,
				"name": "Test Event"
			}`,
			scrubber: shutter.ScrubUnixTimestamp(),
			title:    "Scrubbed Unix Timestamps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutter.SnapJSON(t, tt.title, tt.json, tt.scrubber)
		})
	}
}

func TestCustomScrubbers(t *testing.T) {
	t.Run("regex_scrubber", func(t *testing.T) {
		jsonStr := `{
			"api_key": "sk_live_abc123def456",
			"secret_key": "sk_test_xyz789uvw012",
			"name": "Test User"
		}`

		shutter.SnapJSON(t, "Custom Regex Scrubber", jsonStr,
			shutter.ScrubRegex(`sk_(live|test)_[a-zA-Z0-9]+`, "<API_KEY>"),
		)
	})

	t.Run("exact_match_scrubber", func(t *testing.T) {
		content := "The secret password is 'p@ssw0rd123' and should be hidden."

		shutter.SnapString(t, "Exact Match Scrubber", content,
			shutter.ScrubExact("p@ssw0rd123", "<PASSWORD>"),
		)
	})

	t.Run("custom_function_scrubber", func(t *testing.T) {
		content := "Hello World! This is a TEST."

		shutter.SnapString(t, "Custom Scrubber", content,
			shutter.ScrubWith(func(s string) string {
				return strings.ToLower(s)
			}),
		)
	})
}

func TestScrubWithSnapFunction(t *testing.T) {
	data := map[string]any{
		"user_id":    "550e8400-e29b-41d4-a716-446655440000",
		"email":      "user@example.com",
		"created_at": "2023-01-15T10:30:00Z",
		"name":       "John Doe",
	}

	shutter.Snap(t, "Scrub With Snap", data,
		shutter.ScrubUUID(),
		shutter.ScrubEmail(),
		shutter.ScrubTimestamp(),
	)
}
