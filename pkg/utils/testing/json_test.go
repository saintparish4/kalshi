package testing

import (
	"testing"

	"kalshi/pkg/utils"
)

func TestPrettyJSON(t *testing.T) {
	data := map[string]interface{}{
		"name": "John",
		"age":  30,
		"city": "New York",
	}

	pretty, err := utils.PrettyJSON(data)
	if err != nil {
		t.Errorf("PrettyJSON() unexpected error: %v", err)
	}

	if pretty == "" {
		t.Error("PrettyJSON() returned empty string")
	}

	// Should contain the data
	if !contains(pretty, "John") || !contains(pretty, "30") || !contains(pretty, "New York") {
		t.Errorf("PrettyJSON() output = %v, should contain data", pretty)
	}

	// Should have proper indentation (contain spaces)
	if !contains(pretty, "  ") {
		t.Error("PrettyJSON() should have proper indentation")
	}
}

func TestCompactJSON(t *testing.T) {
	jsonStr := `{
		"name": "John",
		"age": 30
	}`

	compact, err := utils.CompactJSON(jsonStr)
	if err != nil {
		t.Errorf("CompactJSON() unexpected error: %v", err)
	}

	if compact == "" {
		t.Error("CompactJSON() returned empty string")
	}

	// Should not contain newlines or extra spaces
	if contains(compact, "\n") || contains(compact, "  ") {
		t.Errorf("CompactJSON() should not contain newlines or extra spaces: %v", compact)
	}

	// Should still contain the data
	if !contains(compact, "John") || !contains(compact, "30") {
		t.Errorf("CompactJSON() output = %v, should contain data", compact)
	}
}

func TestJSONEqual(t *testing.T) {
	json1 := `{"name": "John", "age": 30}`
	json2 := `{"age": 30, "name": "John"}` // Same data, different order
	json3 := `{"name": "John", "age": 25}` // Different data

	// Test equal JSONs
	if !utils.JSONEqual(json1, json2) {
		t.Error("JSONEqual() should return true for equivalent JSONs")
	}

	// Test different JSONs
	if utils.JSONEqual(json1, json3) {
		t.Error("JSONEqual() should return false for different JSONs")
	}

	// Test invalid JSON
	if utils.JSONEqual(json1, "invalid json") {
		t.Error("JSONEqual() should return false for invalid JSON")
	}
}

func TestExtractJSONField(t *testing.T) {
	jsonStr := `{
		"user": {
			"name": "John",
			"profile": {
				"age": 30,
				"city": "New York"
			}
		}
	}`

	// Test simple field
	result, err := utils.ExtractJSONField(jsonStr, "user.name")
	if err != nil {
		t.Errorf("ExtractJSONField() unexpected error: %v", err)
	}
	if result != "John" {
		t.Errorf("ExtractJSONField() = %v, want %v", result, "John")
	}

	// Test nested field
	result, err = utils.ExtractJSONField(jsonStr, "user.profile.age")
	if err != nil {
		t.Errorf("ExtractJSONField() unexpected error: %v", err)
	}
	if result.(float64) != 30 {
		t.Errorf("ExtractJSONField() = %v, want %v", result, 30)
	}

	// Test non-existent field
	_, err = utils.ExtractJSONField(jsonStr, "user.nonexistent")
	if err == nil {
		t.Error("ExtractJSONField() should return error for non-existent field")
	}

	// Test invalid JSON
	_, err = utils.ExtractJSONField("invalid json", "field")
	if err == nil {
		t.Error("ExtractJSONField() should return error for invalid JSON")
	}
}

func TestMergeJSON(t *testing.T) {
	json1 := `{"name": "John", "age": 30}`
	json2 := `{"city": "New York", "country": "USA"}`
	json3 := `{"age": 25, "job": "Developer"}`

	merged, err := utils.MergeJSON(json1, json2, json3)
	if err != nil {
		t.Errorf("MergeJSON() unexpected error: %v", err)
	}

	if merged == "" {
		t.Error("MergeJSON() returned empty string")
	}

	// Should contain all fields
	if !contains(merged, "John") || !contains(merged, "New York") || !contains(merged, "Developer") {
		t.Errorf("MergeJSON() output = %v, should contain all fields", merged)
	}

	// Later JSONs should override earlier ones
	if !contains(merged, "25") && contains(merged, "30") {
		t.Error("MergeJSON() should override duplicate fields with later values")
	}
}

func TestFilterJSONFields(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30, "city": "New York", "country": "USA"}`
	fields := []string{"name", "age"}

	filtered, err := utils.FilterJSONFields(jsonStr, fields)
	if err != nil {
		t.Errorf("FilterJSONFields() unexpected error: %v", err)
	}

	if filtered == "" {
		t.Error("FilterJSONFields() returned empty string")
	}

	// Should contain only specified fields
	if !contains(filtered, "John") || !contains(filtered, "30") {
		t.Errorf("FilterJSONFields() output = %v, should contain specified fields", filtered)
	}

	// Should not contain other fields
	if contains(filtered, "New York") || contains(filtered, "USA") {
		t.Errorf("FilterJSONFields() output = %v, should not contain other fields", filtered)
	}
}

func TestExcludeJSONFields(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30, "city": "New York", "country": "USA"}`
	fields := []string{"age", "country"}

	excluded, err := utils.ExcludeJSONFields(jsonStr, fields)
	if err != nil {
		t.Errorf("ExcludeJSONFields() unexpected error: %v", err)
	}

	if excluded == "" {
		t.Error("ExcludeJSONFields() returned empty string")
	}

	// Should contain non-excluded fields
	if !contains(excluded, "John") || !contains(excluded, "New York") {
		t.Errorf("ExcludeJSONFields() output = %v, should contain non-excluded fields", excluded)
	}

	// Should not contain excluded fields
	if contains(excluded, "30") || contains(excluded, "USA") {
		t.Errorf("ExcludeJSONFields() output = %v, should not contain excluded fields", excluded)
	}
}

func TestJSONToQuery(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30, "city": "New York"}`

	query, err := utils.JSONToQuery(jsonStr)
	if err != nil {
		t.Errorf("JSONToQuery() unexpected error: %v", err)
	}

	if query == "" {
		t.Error("JSONToQuery() returned empty string")
	}

	// Should contain key-value pairs
	if !contains(query, "name=John") || !contains(query, "age=30") || !contains(query, "city=New York") {
		t.Errorf("JSONToQuery() output = %v, should contain key-value pairs", query)
	}

	// Should contain ampersands
	if !contains(query, "&") {
		t.Error("JSONToQuery() should contain ampersands to separate parameters")
	}
}

func TestQueryToJSON(t *testing.T) {
	query := "name=John&age=30&city=New%20York"

	json, err := utils.QueryToJSON(query)
	if err != nil {
		t.Errorf("QueryToJSON() unexpected error: %v", err)
	}

	if json == "" {
		t.Error("QueryToJSON() returned empty string")
	}

	// Should contain the data (URL encoding is expected)
	if !contains(json, "John") || !contains(json, "30") || !contains(json, "New%20York") {
		t.Errorf("QueryToJSON() output = %v, should contain data", json)
	}

	// Test empty query
	json, err = utils.QueryToJSON("")
	if err != nil {
		t.Errorf("QueryToJSON() with empty query unexpected error: %v", err)
	}

	if json != "{}" {
		t.Errorf("QueryToJSON() with empty query = %v, want %v", json, "{}")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
