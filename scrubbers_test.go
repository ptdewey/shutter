package freeze_test

import (
	"strings"
	"testing"

	"github.com/ptdewey/freeze"
)

func TestScrubUUIDs(t *testing.T) {
	jsonStr := `{
		"user_id": "550e8400-e29b-41d4-a716-446655440000",
		"session_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"name": "John Doe"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed UUIDs", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubUUIDs(),
	})
}

func TestScrubTimestamps(t *testing.T) {
	jsonStr := `{
		"created_at": "2023-01-15T10:30:00Z",
		"updated_at": "2023-11-20T15:45:30.123Z",
		"deleted_at": "2023-12-01T08:00:00+05:00",
		"name": "Test Event"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed Timestamps", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubTimestamps(),
	})
}

func TestScrubEmails(t *testing.T) {
	jsonStr := `{
		"email": "user@example.com",
		"backup_email": "backup.user+tag@subdomain.example.co.uk",
		"name": "John Doe"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed Emails", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubEmails(),
	})
}

func TestScrubIPAddresses(t *testing.T) {
	jsonStr := `{
		"client_ip": "192.168.1.1",
		"server_ip": "10.0.0.5",
		"message": "Connection from 172.16.0.100"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed IPs", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubIPAddresses(),
	})
}

func TestScrubJWTs(t *testing.T) {
	jsonStr := `{
		"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed JWTs", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubJWTs(),
	})
}

func TestMultipleScrubbers(t *testing.T) {
	jsonStr := `{
		"user_id": "550e8400-e29b-41d4-a716-446655440000",
		"email": "user@example.com",
		"created_at": "2023-01-15T10:30:00Z",
		"ip_address": "192.168.1.1",
		"name": "John Doe"
	}`

	freeze.SnapJSONWithOptions(t, "Multiple Scrubbers", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubUUIDs(),
		freeze.ScrubEmails(),
		freeze.ScrubTimestamps(),
		freeze.ScrubIPAddresses(),
	})
}

func TestRegexScrubber(t *testing.T) {
	jsonStr := `{
		"api_key": "sk_live_abc123def456",
		"secret_key": "sk_test_xyz789uvw012",
		"name": "Test User"
	}`

	freeze.SnapJSONWithOptions(t, "Custom Regex Scrubber", jsonStr, []freeze.SnapshotOption{
		freeze.RegexScrubber(`sk_(live|test)_[a-zA-Z0-9]+`, "<API_KEY>"),
	})
}

func TestExactMatchScrubber(t *testing.T) {
	content := "The secret password is 'p@ssw0rd123' and should be hidden."

	freeze.SnapStringWithOptions(t, "Exact Match Scrubber", content, []freeze.SnapshotOption{
		freeze.ExactMatchScrubber("p@ssw0rd123", "<PASSWORD>"),
	})
}

func TestCustomScrubber(t *testing.T) {
	content := "Hello World! This is a TEST."

	freeze.SnapStringWithOptions(t, "Custom Scrubber", content, []freeze.SnapshotOption{
		freeze.CustomScrubber(func(s string) string {
			return strings.ToLower(s)
		}),
	})
}

func TestScrubDates(t *testing.T) {
	jsonStr := `{
		"birth_date": "1990-05-15",
		"hire_date": "2020-01-01",
		"us_format": "12/25/2023",
		"name": "John Doe"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed Dates", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubDates(),
	})
}

func TestScrubAPIKeys(t *testing.T) {
	jsonStr := `{
		"stripe_key": "sk_live_51HqZ2bKl4FGBMFpLxO0123",
		"test_key": "pk_test_51HqZ2bKl4FGBMFpLxO0456",
		"api_key_prod": "api_key_prod_abc123def456",
		"name": "Test Config"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed API Keys", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubAPIKeys(),
	})
}

func TestScrubWithSnapFunction(t *testing.T) {
	data := map[string]any{
		"user_id":    "550e8400-e29b-41d4-a716-446655440000",
		"email":      "user@example.com",
		"created_at": "2023-01-15T10:30:00Z",
		"name":       "John Doe",
	}

	freeze.SnapWithOptions(t, "Scrub With Snap", []freeze.SnapshotOption{
		freeze.ScrubUUIDs(),
		freeze.ScrubEmails(),
		freeze.ScrubTimestamps(),
	}, data)
}

func TestCreditCardScrubbing(t *testing.T) {
	jsonStr := `{
		"card_number": "4532-1234-5678-9010",
		"backup_card": "4532 1234 5678 9010",
		"another_card": "4532123456789010",
		"name": "John Doe"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed Credit Cards", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubCreditCards(),
	})
}

func TestUnixTimestampScrubbing(t *testing.T) {
	jsonStr := `{
		"created": 1699999999,
		"updated": 1700000000000,
		"deleted": 1700000000,
		"name": "Test Event"
	}`

	freeze.SnapJSONWithOptions(t, "Scrubbed Unix Timestamps", jsonStr, []freeze.SnapshotOption{
		freeze.ScrubUnixTimestamps(),
	})
}
