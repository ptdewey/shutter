package shutter_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ptdewey/shutter"
)

func TestSnapMultiple(t *testing.T) {
	shutter.SnapMany(t, "Multiple Values Test", []any{"value1", "value2", 42, "foo", "bar", "baz", "wibble", "wobble", "tock", nil})
}

type CustomStruct struct {
	Name string
	Age  int
}

func (c CustomStruct) Format() string {
	return fmt.Sprintf("CustomStruct{Name: %s, Age: %d}", c.Name, c.Age)
}

func TestSnapCustomType(t *testing.T) {
	cs := CustomStruct{
		Name: "Alice",
		Age:  30,
	}
	shutter.Snap(t, "Custom Type Test", cs)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func cleanupTestSnapshots(t *testing.T) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Logf("failed to get cwd: %v", err)
		return
	}

	snapshotDir := filepath.Join(cwd, "__snapshots__")
	_ = os.RemoveAll(snapshotDir)
}

// ============================================================================
// COMPLEX GO STRUCTURES TESTS
// ============================================================================

type User struct {
	ID        int
	Username  string
	Email     string
	Active    bool
	CreatedAt time.Time
	Roles     []string
	Metadata  map[string]any
}

type Post struct {
	ID        int
	Title     string
	Content   string
	Author    User
	Tags      []string
	Comments  []Comment
	Likes     int
	Published bool
	CreatedAt time.Time
}

type Comment struct {
	ID        int
	Author    string
	Content   string
	CreatedAt time.Time
	Replies   []Comment
}

func TestComplexNestedStructure(t *testing.T) {
	user := User{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		Active:    true,
		CreatedAt: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		Roles:     []string{"admin", "moderator", "user"},
		Metadata: map[string]any{
			"theme":         "dark",
			"notifications": true,
			"language":      "en",
			"preferences": map[string]any{
				"email_frequency": "weekly",
				"notifications":   true,
			},
		},
	}

	comments := []Comment{
		{
			ID:        1,
			Author:    "alice",
			Content:   "Great post!",
			CreatedAt: time.Date(2023, 2, 1, 14, 22, 0, 0, time.UTC),
			Replies: []Comment{
				{
					ID:        2,
					Author:    "bob",
					Content:   "I agree!",
					CreatedAt: time.Date(2023, 2, 1, 15, 45, 0, 0, time.UTC),
					Replies:   []Comment{},
				},
			},
		},
		{
			ID:        3,
			Author:    "charlie",
			Content:   "Thanks for sharing!",
			CreatedAt: time.Date(2023, 2, 2, 9, 30, 0, 0, time.UTC),
			Replies:   []Comment{},
		},
	}

	post := Post{
		ID:        100,
		Title:     "Introduction to Go Snapshot Testing",
		Content:   "This is a comprehensive guide to snapshot testing in Go...",
		Author:    user,
		Tags:      []string{"go", "testing", "snapshots", "best-practices"},
		Comments:  comments,
		Likes:     42,
		Published: true,
		CreatedAt: time.Date(2023, 1, 20, 9, 0, 0, 0, time.UTC),
	}

	shutter.Snap(t, "Complex Nested Structure", post)
}

func TestMultipleComplexStructures(t *testing.T) {
	users := []User{
		{
			ID:       1,
			Username: "alice",
			Email:    "alice@example.com",
			Active:   true,
			Roles:    []string{"user", "moderator"},
			Metadata: map[string]any{
				"verified": true,
				"badge":    "verified",
			},
		},
		{
			ID:       2,
			Username: "bob",
			Email:    "bob@example.com",
			Active:   false,
			Roles:    []string{"user"},
			Metadata: map[string]any{
				"verified": false,
				"avatar":   "https://example.com/bob.jpg",
			},
		},
		{
			ID:       3,
			Username: "charlie",
			Email:    "charlie@example.com",
			Active:   true,
			Roles:    []string{"user", "admin"},
			Metadata: map[string]any{
				"verified":         true,
				"account_age_days": 365,
			},
		},
	}

	shutter.Snap(t, "Multiple Complex Structures", users)
}

func TestStructureWithInterface(t *testing.T) {
	type Response struct {
		Status  string
		Message string
		Data    any
		Meta    map[string]any
	}

	responses := []Response{
		{
			Status:  "success",
			Message: "User retrieved",
			Data: User{
				ID:       1,
				Username: "john",
				Email:    "john@example.com",
				Active:   true,
			},
			Meta: map[string]any{
				"request_id": "req-123",
				"timestamp":  "2023-01-20T10:30:00Z",
			},
		},
		{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
			Meta: map[string]any{
				"error_code": 404,
				"error_type": "NOT_FOUND",
			},
		},
		{
			Status:  "success",
			Message: "Posts retrieved",
			Data: []Post{
				{
					ID:        1,
					Title:     "First Post",
					Published: true,
				},
			},
			Meta: map[string]any{
				"total_count": 10,
				"page":        1,
				"per_page":    20,
			},
		},
	}

	shutter.Snap(t, "Structure with Interface Fields", responses)
}

func TestNestedMapsAndSlices(t *testing.T) {
	complexData := map[string]any{
		"users": map[string]any{
			"active": []map[string]any{
				{
					"id":       1,
					"name":     "Alice",
					"verified": true,
				},
				{
					"id":       2,
					"name":     "Bob",
					"verified": false,
				},
			},
			"inactive": []map[string]any{
				{
					"id":   3,
					"name": "Charlie",
				},
			},
		},
		"posts": map[string]any{
			"published":  42,
			"drafts":     5,
			"categories": []string{"tech", "lifestyle", "news"},
		},
		"stats": map[string]any{
			"daily": map[string]any{
				"views":  1500,
				"clicks": 320,
				"conversions": map[string]any{
					"total": 45,
					"by_source": map[string]int{
						"organic":  25,
						"paid":     15,
						"referral": 5,
					},
				},
			},
		},
	}

	shutter.Snap(t, "Nested Maps and Slices", complexData)
}

func TestStructureWithPointers(t *testing.T) {
	type Address struct {
		Street string
		City   string
		Zip    string
	}

	type Person struct {
		Name    string
		Age     int
		Address *Address
		Manager *Person
		Friends []*Person
		Email   *string
	}

	addr1 := &Address{
		Street: "123 Main St",
		City:   "Boston",
		Zip:    "02101",
	}

	email := "jane@example.com"

	person1 := Person{
		Name:    "Jane",
		Age:     30,
		Address: addr1,
		Email:   &email,
	}

	person2 := Person{
		Name:    "John",
		Age:     35,
		Address: addr1,
		Manager: &person1,
		Friends: []*Person{&person1},
	}

	shutter.Snap(t, "Structure with Pointers", person2)
}

func TestStructureWithEmptyValues(t *testing.T) {
	type Container struct {
		Items      []string
		Tags       map[string]string
		OptionalID *int
		Count      int
		Active     bool
	}

	containers := []Container{
		{
			Items:      []string{},
			Tags:       map[string]string{},
			OptionalID: nil,
			Count:      0,
			Active:     false,
		},
		{
			Items:      nil,
			Tags:       nil,
			OptionalID: nil,
			Count:      0,
			Active:     true,
		},
		{
			Items:      []string{"a", "b", "c"},
			Tags:       map[string]string{"type": "test", "env": "dev"},
			OptionalID: ptr(42),
			Count:      3,
			Active:     true,
		},
	}

	shutter.Snap(t, "Structure with Empty Values", containers)
}

// ============================================================================
// JSON TESTS - Focus on edge cases and special handling
// ============================================================================

func TestJsonWithSpecialCharacters(t *testing.T) {
	jsonStr := `{
		"english": "Hello, World!",
		"unicode": "こんにちは 世界 🌍",
		"emoji": "😀 😃 😄 😁 😆",
		"special_chars": "!@#$%^&*()_+-=[]{}|;:,.<>?",
		"escaped": "quotes: \"double\" and 'single'",
		"newlines": "line1\nline2\rline3\r\nline4",
		"tabs": "col1\tcol2\tcol3",
		"backslash": "path\\to\\file"
	}`

	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	shutter.Snap(t, "JSON with Special Characters", data)
}

func TestLargeJson(t *testing.T) {
	type Product struct {
		ID          int      `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		Stock       int      `json:"stock"`
		InStock     bool     `json:"in_stock"`
		Tags        []string `json:"tags"`
	}

	type Order struct {
		ID          int        `json:"id"`
		CustomerID  int        `json:"customer_id"`
		Products    []Product  `json:"products"`
		Total       float64    `json:"total"`
		Status      string     `json:"status"`
		CreatedAt   time.Time  `json:"created_at"`
		ShippedAt   *time.Time `json:"shipped_at"`
		DeliveredAt *time.Time `json:"delivered_at"`
	}

	shippedTime := time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC)

	order := Order{
		ID:         1001,
		CustomerID: 42,
		Products: []Product{
			{
				ID:          1,
				Name:        "Laptop",
				Description: "High-performance laptop",
				Price:       999.99,
				Stock:       5,
				InStock:     true,
				Tags:        []string{"electronics", "computers", "laptops"},
			},
			{
				ID:          2,
				Name:        "Mouse",
				Description: "Wireless mouse",
				Price:       29.99,
				Stock:       50,
				InStock:     true,
				Tags:        []string{"electronics", "accessories"},
			},
			{
				ID:          3,
				Name:        "Keyboard",
				Description: "Mechanical keyboard",
				Price:       149.99,
				Stock:       0,
				InStock:     false,
				Tags:        []string{"electronics", "accessories"},
			},
		},
		Total:       1179.97,
		Status:      "shipped",
		CreatedAt:   time.Date(2023, 1, 28, 14, 30, 0, 0, time.UTC),
		ShippedAt:   &shippedTime,
		DeliveredAt: nil,
	}

	jsonBytes, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}

	var data any
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	shutter.Snap(t, "Large JSON Structure", data)
}

// ============================================================================
// SNAPJSON FUNCTION TESTS - Focus on edge cases and real-world examples
// ============================================================================

func TestSnapJsonWithNestedObjects(t *testing.T) {
	jsonStr := `{
		"user": {
			"id": 42,
			"profile": {
				"username": "jane_smith",
				"avatar": "https://example.com/avatar.jpg",
				"settings": {
					"theme": "dark",
					"notifications": true,
					"language": "en"
				}
			},
			"permissions": ["read", "write", "admin"]
		},
		"created_at": "2023-06-15T10:30:00Z"
	}`

	shutter.SnapJSON(t, "SnapJSON Nested Objects", jsonStr)
}

func TestSnapJsonComplexAPI(t *testing.T) {
	jsonStr := `{
		"status": "success",
		"code": 200,
		"data": {
			"users": [
				{
					"id": 1,
					"name": "Alice",
					"role": "admin",
					"department": "Engineering",
					"active": true
				},
				{
					"id": 2,
					"name": "Bob",
					"role": "user",
					"department": "Sales",
					"active": true
				},
				{
					"id": 3,
					"name": "Charlie",
					"role": "user",
					"department": "Marketing",
					"active": false
				}
			],
			"pagination": {
				"page": 1,
				"per_page": 10,
				"total": 3,
				"total_pages": 1
			}
		},
		"timestamp": "2023-11-18T21:45:30Z"
	}`

	shutter.SnapJSON(t, "SnapJSON Complex API Response", jsonStr)
}

func TestSnapJsonMixedTypes(t *testing.T) {
	jsonStr := `{
		"mixed_array": [
			"string",
			123,
			45.67,
			true,
			false,
			null,
			{"nested": "object"},
			[1, 2, 3]
		],
		"complex": [
			{"type": "user", "id": 1},
			{"type": "post", "id": 100},
			[1, 2, 3],
			"string",
			null
		]
	}`

	shutter.SnapJSON(t, "SnapJSON Mixed Types", jsonStr)
}

func TestSnapJsonRealWorldExample(t *testing.T) {
	jsonStr := `{
		"success": true,
		"data": {
			"product": {
				"id": "prod_12345",
				"name": "Premium Wireless Headphones",
				"sku": "PWH-001",
				"description": "High-quality wireless headphones with noise cancellation",
				"price": {
					"amount": 199.99,
					"currency": "USD",
					"discount": 10,
					"final_price": 179.99
				},
				"inventory": {
					"total": 500,
					"available": 425,
					"reserved": 50,
					"damaged": 25
				},
				"specifications": {
					"battery_life": "30 hours",
					"weight": "250g",
					"colors": ["black", "white", "blue"],
					"warranty_months": 24
				},
				"ratings": {
					"average": 4.5,
					"count": 1250,
					"breakdown": {
						"5": 750,
						"4": 350,
						"3": 100,
						"2": 30,
						"1": 20
					}
				},
				"reviews": [
					{
						"id": "rev_001",
						"user": "john_doe",
						"rating": 5,
						"title": "Excellent product!",
						"content": "Great sound quality and comfortable to wear.",
						"helpful": 25,
						"created_at": "2023-11-15T10:30:00Z"
					},
					{
						"id": "rev_002",
						"user": "jane_smith",
						"rating": 4,
						"title": "Good but pricey",
						"content": "Works well, could be cheaper.",
						"helpful": 12,
						"created_at": "2023-11-10T14:20:00Z"
					}
				]
			},
			"related_products": [
				{
					"id": "prod_12346",
					"name": "Headphone Case",
					"price": 29.99
				},
				{
					"id": "prod_12347",
					"name": "Audio Cable",
					"price": 14.99
				}
			]
		},
		"request_id": "req_abc123def456",
		"timestamp": "2023-11-18T22:00:00Z"
	}`

	shutter.SnapJSON(t, "SnapJSON Real World Example", jsonStr)
}

func ptr[T any](t T) *T { return &t }
