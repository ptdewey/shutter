package freeze_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ptdewey/freeze"
	"github.com/ptdewey/freeze/internal/api"
)

func TestSnapString(t *testing.T) {
	freeze.SnapString(t, "Simple String Test", "hello world")
}

func TestSnapMultiple(t *testing.T) {
	freeze.Snap(t, "Multiple Values Test", "value1", "value2", 42, "foo", "bar", "baz", "wibble", "wobble", "tock", nil)
}

type CustomStruct struct {
	Name string
	Age  int
}

func (c CustomStruct) Format() string {
	return "CustomStruct{Name: " + c.Name + ", Age: " + string(rune(c.Age)) + "}"
}

func TestSnapCustomType(t *testing.T) {
	cs := CustomStruct{
		Name: "Alice",
		Age:  30,
	}
	freeze.Snap(t, "Custom Type Test", cs)
}

func TestMap(t *testing.T) {
	freeze.Snap(t, "Map Test", map[string]any{
		"foo":    "bar",
		"wibble": "wobble",
	})
}

func TestSerializeDeserialize(t *testing.T) {
	snap := &freeze.Snapshot{
		Title:   "My Test Title",
		Name:    "TestExample",
		Content: "test content\nmultiline",
	}

	serialized := snap.Serialize()
	expected := "---\ntitle: My Test Title\ntest_name: TestExample\nfile_path: \nfunc_name: \nversion: \n---\ntest content\nmultiline"
	if serialized != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, serialized)
	}

	deserialized, err := freeze.Deserialize(serialized)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	if deserialized.Title != snap.Title {
		t.Errorf("title mismatch: %s != %s", deserialized.Title, snap.Title)
	}
	if deserialized.Name != snap.Name {
		t.Errorf("test name mismatch: %s != %s", deserialized.Name, snap.Name)
	}
	if deserialized.Content != snap.Content {
		t.Errorf("content mismatch: %s != %s", deserialized.Content, snap.Content)
	}
}

func TestFileOperations(t *testing.T) {
	snap := &freeze.Snapshot{
		Title:   "File Ops Title",
		Name:    "TestFileOps",
		Content: "file test content",
	}

	if err := api.SaveSnapshot(snap, "test"); err != nil {
		t.Fatalf("failed to save snapshot: %v", err)
	}

	read, err := freeze.ReadSnapshot("TestFileOps", "test")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	if read.Content != snap.Content {
		t.Errorf("content mismatch: %s != %s", read.Content, snap.Content)
	}

	// cleanupTestSnapshots(t)
}

func TestSnapshotFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestMyFunction", "test_my_function"},
		{"test_another_one", "test_another_one"},
		{"TestCamelCase", "test_camel_case"},
		{"TestWithNumbers123", "test_with_numbers123"},
	}

	for _, tt := range tests {
		result := freeze.SnapshotFileName(tt.input)
		if result != tt.expected {
			t.Errorf("SnapshotFileName(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestHistogramDiff(t *testing.T) {
	oldStr := "line1\nline2\nline3"
	newStr := "line1\nmodified\nline3"

	diff := freeze.Histogram(oldStr, newStr)

	if len(diff) < 3 {
		t.Errorf("expected at least 3 diff lines, got %d", len(diff))
	}

	if diff[0].Kind != freeze.DiffShared || diff[0].Line != "line1" {
		t.Errorf("line 0: expected shared 'line1', got %v %s", diff[0].Kind, diff[0].Line)
	}

	hasModified := false
	for _, d := range diff {
		if d.Line == "modified" {
			hasModified = true
			if d.Kind != freeze.DiffNew {
				t.Errorf("'modified' should be marked as new")
			}
		}
	}
	if !hasModified {
		t.Error("diff missing 'modified' line")
	}

	hasLine3 := false
	for _, d := range diff {
		if d.Line == "line3" && d.Kind == freeze.DiffShared {
			hasLine3 = true
		}
	}
	if !hasLine3 {
		t.Error("diff should have 'line3' as shared")
	}
}

func TestDiffSnapshotBox(t *testing.T) {
	old := &freeze.Snapshot{
		Title:   "Diff Test Title",
		Name:    "TestDiff",
		Content: "old content",
	}

	new := &freeze.Snapshot{
		Title:   "Diff Test Title",
		Name:    "TestDiff",
		Content: "new content",
	}

	box := freeze.DiffSnapshotBox(old, new)
	if box == "" {
		t.Error("DiffSnapshotBox returned empty string")
	}

	if !contains(box, "Snapshot Diff") {
		t.Error("DiffSnapshotBox missing header")
	}
}

func TestNewSnapshotBox(t *testing.T) {
	snap := &freeze.Snapshot{
		Title:   "New Test Title",
		Name:    "TestNew",
		Content: "test content",
	}

	box := freeze.NewSnapshotBox(snap)
	if box == "" {
		t.Error("NewSnapshotBox returned empty string")
	}

	if !contains(box, "New Snapshot") {
		t.Error("NewSnapshotBox missing header")
	}
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

// User represents a user in a system
type User struct {
	ID        int
	Username  string
	Email     string
	Active    bool
	CreatedAt time.Time
	Roles     []string
	Metadata  map[string]interface{}
}

// Post represents a blog post
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

// Comment represents a comment on a post
type Comment struct {
	ID        int
	Author    string
	Content   string
	CreatedAt time.Time
	Replies   []Comment
}

// TestComplexNestedStructure tests snapshot with deeply nested Go structures
func TestComplexNestedStructure(t *testing.T) {
	user := User{
		ID:        1,
		Username:  "john_doe",
		Email:     "john@example.com",
		Active:    true,
		CreatedAt: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
		Roles:     []string{"admin", "moderator", "user"},
		Metadata: map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"language":      "en",
			"preferences": map[string]interface{}{
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

	freeze.Snap(t, "Complex Nested Structure", post)
}

// TestMultipleComplexStructures tests snapshot with multiple complex structures
func TestMultipleComplexStructures(t *testing.T) {
	users := []User{
		{
			ID:       1,
			Username: "alice",
			Email:    "alice@example.com",
			Active:   true,
			Roles:    []string{"user", "moderator"},
			Metadata: map[string]interface{}{
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
			Metadata: map[string]interface{}{
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
			Metadata: map[string]interface{}{
				"verified":         true,
				"account_age_days": 365,
			},
		},
	}

	freeze.Snap(t, "Multiple Complex Structures", users)
}

// TestStructureWithInterface tests structures containing interface{} fields
func TestStructureWithInterface(t *testing.T) {
	type Response struct {
		Status  string
		Message string
		Data    interface{}
		Meta    map[string]interface{}
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
			Meta: map[string]interface{}{
				"request_id": "req-123",
				"timestamp":  "2023-01-20T10:30:00Z",
			},
		},
		{
			Status:  "error",
			Message: "User not found",
			Data:    nil,
			Meta: map[string]interface{}{
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
			Meta: map[string]interface{}{
				"total_count": 10,
				"page":        1,
				"per_page":    20,
			},
		},
	}

	freeze.Snap(t, "Structure with Interface Fields", responses)
}

// TestNestedMapsAndSlices tests complex nested maps and slices
func TestNestedMapsAndSlices(t *testing.T) {
	complexData := map[string]interface{}{
		"users": map[string]interface{}{
			"active": []map[string]interface{}{
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
			"inactive": []map[string]interface{}{
				{
					"id":   3,
					"name": "Charlie",
				},
			},
		},
		"posts": map[string]interface{}{
			"published":  42,
			"drafts":     5,
			"categories": []string{"tech", "lifestyle", "news"},
		},
		"stats": map[string]interface{}{
			"daily": map[string]interface{}{
				"views":  1500,
				"clicks": 320,
				"conversions": map[string]interface{}{
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

	freeze.Snap(t, "Nested Maps and Slices", complexData)
}

// TestStructureWithPointers tests structures with pointer fields
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

	freeze.Snap(t, "Structure with Pointers", person2)
}

// TestStructureWithEmptyValues tests structures with empty slices, maps, nil values
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
			OptionalID: intPtr(42),
			Count:      3,
			Active:     true,
		},
	}

	freeze.Snap(t, "Structure with Empty Values", containers)
}

// ============================================================================
// JSON OBJECT TESTS
// ============================================================================

// TestJSONObject tests snapshot with JSON objects
func TestJSONObject(t *testing.T) {
	jsonStr := `{
		"user": {
			"id": 1,
			"username": "john_doe",
			"email": "john@example.com",
			"profile": {
				"first_name": "John",
				"last_name": "Doe",
				"bio": "Software engineer",
				"verified": true
			},
			"roles": ["user", "admin"],
			"created_at": "2023-01-15T10:30:00Z"
		},
		"status": "success",
		"message": null
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "JSON Object", data)
}

// TestComplexJSONStructure tests complex nested JSON structures
func TestComplexJSONStructure(t *testing.T) {
	jsonStr := `{
		"api": {
			"version": "2.0",
			"endpoints": [
				{
					"path": "/users",
					"method": "GET",
					"auth_required": true,
					"rate_limit": {
						"requests": 100,
						"window": "1m"
					},
					"responses": {
						"200": {
							"description": "Success",
							"schema": {
								"type": "array",
								"items": {
									"type": "object",
									"properties": {
										"id": {"type": "integer"},
										"name": {"type": "string"}
									}
								}
							}
						},
						"401": {
							"description": "Unauthorized"
						}
					}
				},
				{
					"path": "/users/{id}",
					"method": "POST",
					"auth_required": true,
					"rate_limit": {
						"requests": 50,
						"window": "1m"
					}
				}
			],
			"models": {
				"User": {
					"properties": {
						"id": {"type": "integer"},
						"username": {"type": "string"},
						"email": {"type": "string", "format": "email"},
						"created_at": {"type": "string", "format": "date-time"},
						"roles": {
							"type": "array",
							"items": {"type": "string"}
						}
					},
					"required": ["id", "username", "email"]
				}
			}
		}
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "Complex JSON Structure", data)
}

// TestJSONArrayOfObjects tests JSON arrays with multiple object types
func TestJSONArrayOfObjects(t *testing.T) {
	jsonStr := `[
		{
			"type": "user",
			"id": 1,
			"data": {
				"name": "Alice",
				"role": "admin"
			}
		},
		{
			"type": "post",
			"id": 100,
			"data": {
				"title": "First Post",
				"author_id": 1,
				"likes": 42
			}
		},
		{
			"type": "comment",
			"id": 500,
			"data": {
				"content": "Great post!",
				"author_id": 2,
				"post_id": 100
			}
		}
	]`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "JSON Array of Objects", data)
}

// TestJSONWithVariousTypes tests JSON with various data types
func TestJSONWithVariousTypes(t *testing.T) {
	jsonStr := `{
		"string": "hello world",
		"integer": 42,
		"float": 3.14159,
		"boolean_true": true,
		"boolean_false": false,
		"null_value": null,
		"array": [1, 2, 3, "four", 5.5],
		"object": {
			"nested": "value",
			"count": 10
		},
		"empty_array": [],
		"empty_object": {},
		"escaped_string": "line1\nline2\ttab"
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "JSON with Various Types", data)
}

// TestJSONNumbers tests JSON with various number formats
func TestJSONNumbers(t *testing.T) {
	jsonStr := `{
		"integers": {
			"zero": 0,
			"positive": 42,
			"negative": -100,
			"large": 9999999999999
		},
		"floats": {
			"small": 0.0001,
			"pi": 3.14159265359,
			"scientific": 1.23e-4,
			"negative_float": -42.5
		},
		"edge_cases": {
			"one": 1,
			"minus_one": -1
		}
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "JSON Numbers", data)
}

// TestJSONWithSpecialCharacters tests JSON with special characters and unicode
func TestJSONWithSpecialCharacters(t *testing.T) {
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

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "JSON with Special Characters", data)
}

// TestGoStructMarshalledToJSON tests Go struct marshalled to JSON
func TestGoStructMarshalledToJSON(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
		Zip    string `json:"zip"`
	}

	type Contact struct {
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		Phone     string    `json:"phone"`
		Address   Address   `json:"address"`
		Tags      []string  `json:"tags"`
		Active    bool      `json:"active"`
		CreatedAt time.Time `json:"created_at"`
	}

	contact := Contact{
		Name:  "Jane Smith",
		Email: "jane@example.com",
		Phone: "+1-555-0123",
		Address: Address{
			Street: "456 Oak Ave",
			City:   "San Francisco",
			Zip:    "94102",
		},
		Tags:      []string{"vip", "verified", "premium"},
		Active:    true,
		CreatedAt: time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC),
	}

	jsonBytes, err := json.MarshalIndent(contact, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}

	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "Go Struct Marshalled to JSON", data)
}

// TestDeeplyNestedJSON tests deeply nested JSON structure
func TestDeeplyNestedJSON(t *testing.T) {
	type Level4 struct {
		Value string
	}

	type Level3 struct {
		L4 Level4
	}

	type Level2 struct {
		L3 Level3
	}

	type Level1 struct {
		L2 Level2
	}

	l1 := Level1{
		L2: Level2{
			L3: Level3{
				L4: Level4{
					Value: "deep value",
				},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(l1, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}

	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "Deeply Nested JSON", data)
}

// TestLargeJSON tests larger JSON structure with many fields
func TestLargeJSON(t *testing.T) {
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

	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "Large JSON Structure", data)
}

// TestJSONWithMixedArrays tests JSON with arrays containing different types
func TestJSONWithMixedArrays(t *testing.T) {
	jsonStr := `{
		"heterogeneous_array": [
			"string",
			42,
			3.14,
			true,
			null,
			{"object": "value"},
			[1, 2, 3]
		],
		"matrix": [
			[1, 2, 3],
			[4, 5, 6],
			[7, 8, 9]
		],
		"object_array": [
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
			{"id": 3, "name": "Item 3"}
		]
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	freeze.Snap(t, "JSON with Mixed Arrays", data)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func intPtr(i int) *int {
	return &i
}
