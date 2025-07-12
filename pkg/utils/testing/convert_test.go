package testing

import (
	"testing"

	"kalshi/pkg/utils"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, ""},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"int8", int8(42), "42"},
		{"int16", int16(42), "42"},
		{"int32", int32(42), "42"},
		{"int64", int64(42), "42"},
		{"uint", uint(42), "42"},
		{"uint8", uint8(42), "42"},
		{"uint16", uint16(42), "42"},
		{"uint32", uint32(42), "42"},
		{"uint64", uint64(42), "42"},
		{"float32", float32(42.5), "42.5"},
		{"float64", 42.5, "42.5"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"bytes", []byte("hello"), "hello"},
		{"struct", struct{ Name string }{"test"}, `{"Name":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.ToString(tt.input); got != tt.expected {
				t.Errorf("ToString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    int
		expectError bool
	}{
		{"int", 42, 42, false},
		{"int8", int8(42), 42, false},
		{"int16", int16(42), 42, false},
		{"int32", int32(42), 42, false},
		{"int64", int64(42), 42, false},
		{"uint", uint(42), 42, false},
		{"uint8", uint8(42), 42, false},
		{"uint16", uint16(42), 42, false},
		{"uint32", uint32(42), 42, false},
		{"uint64", uint64(42), 42, false},
		{"float32", float32(42.5), 42, false},
		{"float64", 42.5, 42, false},
		{"string valid", "42", 42, false},
		{"string invalid", "abc", 0, true},
		{"bool true", true, 1, false},
		{"bool false", false, 0, false},
		{"struct", struct{}{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ToInt(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("ToInt() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ToInt() unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ToInt() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    float64
		expectError bool
	}{
		{"float64", 42.5, 42.5, false},
		{"float32", float32(42.5), 42.5, false},
		{"int", 42, 42.0, false},
		{"int8", int8(42), 42.0, false},
		{"int16", int16(42), 42.0, false},
		{"int32", int32(42), 42.0, false},
		{"int64", int64(42), 42.0, false},
		{"uint", uint(42), 42.0, false},
		{"uint8", uint8(42), 42.0, false},
		{"uint16", uint16(42), 42.0, false},
		{"uint32", uint32(42), 42.0, false},
		{"uint64", uint64(42), 42.0, false},
		{"string valid", "42.5", 42.5, false},
		{"string invalid", "abc", 0, true},
		{"bool true", true, 1.0, false},
		{"bool false", false, 0.0, false},
		{"struct", struct{}{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ToFloat64(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("ToFloat64() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ToFloat64() unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ToFloat64() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    bool
		expectError bool
	}{
		{"bool true", true, true, false},
		{"bool false", false, false, false},
		{"int positive", 42, true, false},
		{"int zero", 0, false, false},
		{"int negative", -1, true, false},
		{"uint positive", uint(42), true, false},
		{"uint zero", uint(0), false, false},
		{"float positive", 42.5, true, false},
		{"float zero", 0.0, false, false},
		{"string true", "true", true, false},
		{"string false", "false", false, false},
		{"string yes", "yes", true, false},
		{"string no", "no", false, false},
		{"string on", "on", true, false},
		{"string off", "off", false, false},
		{"string y", "y", true, false},
		{"string n", "n", false, false},
		{"string t", "t", true, false},
		{"string f", "f", false, false},
		{"string empty", "", false, false},
		{"string invalid", "maybe", false, true},
		{"struct", struct{}{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ToBool(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("ToBool() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ToBool() unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ToBool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    []string
		expectError bool
	}{
		{"nil", nil, []string{}, false},
		{"string slice", []string{"a", "b"}, []string{"a", "b"}, false},
		{"interface slice", []interface{}{"a", 42, true}, []string{"a", "42", "true"}, false},
		{"json array", `["a", "b", "c"]`, []string{"a", "b", "c"}, false},
		{"comma separated", "a,b,c", []string{"a", "b", "c"}, false},
		{"empty string", "", []string{}, false},
		{"invalid type", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ToStringSlice(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("ToStringSlice() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ToStringSlice() unexpected error: %v", err)
			}
			if !tt.expectError {
				if len(got) != len(tt.expected) {
					t.Errorf("ToStringSlice() length = %v, want %v", len(got), len(tt.expected))
				}
				for i, v := range got {
					if v != tt.expected[i] {
						t.Errorf("ToStringSlice()[%d] = %v, want %v", i, v, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestToIntSlice(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    []int
		expectError bool
	}{
		{"nil", nil, []int{}, false},
		{"int slice", []int{1, 2, 3}, []int{1, 2, 3}, false},
		{"interface slice", []interface{}{1, "2", 3}, []int{1, 2, 3}, false},
		{"json array", `[1, 2, 3]`, []int{1, 2, 3}, false},
		{"comma separated", "1,2,3", []int{1, 2, 3}, false},
		{"empty string", "", []int{}, false},
		{"invalid type", "abc", nil, true},
		{"invalid number in string", "1,abc,3", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ToIntSlice(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("ToIntSlice() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ToIntSlice() unexpected error: %v", err)
			}
			if !tt.expectError {
				if len(got) != len(tt.expected) {
					t.Errorf("ToIntSlice() length = %v, want %v", len(got), len(tt.expected))
				}
				for i, v := range got {
					if v != tt.expected[i] {
						t.Errorf("ToIntSlice()[%d] = %v, want %v", i, v, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email,omitempty"`
	}

	testStruct := TestStruct{
		Name: "John",
		Age:  30,
	}

	expected := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	result, err := utils.StructToMap(testStruct)
	if err != nil {
		t.Errorf("StructToMap() unexpected error: %v", err)
	}

	// Debug output
	t.Logf("StructToMap() result: %#v", result)

	if len(result) != len(expected) {
		t.Errorf("StructToMap() length = %v, want %v", len(result), len(expected))
	}

	// Check that the result contains the expected keys and values
	for key, value := range expected {
		gotValue, ok := result[key]
		if !ok {
			t.Errorf("StructToMap() missing key: %s", key)
			continue
		}
		// Handle numeric type differences (int vs float64)
		switch v := value.(type) {
		case int:
			if f, ok := gotValue.(float64); ok {
				if int(f) != v {
					t.Errorf("StructToMap()[%s] = %v, want %v", key, gotValue, value)
				}
				continue
			}
		}
		if gotValue != value {
			t.Errorf("StructToMap()[%s] = %v, want %v", key, gotValue, value)
		}
	}
}

func TestMapToStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	input := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	var result TestStruct
	err := utils.MapToStruct(input, &result)
	if err != nil {
		t.Errorf("MapToStruct() unexpected error: %v", err)
	}

	expected := TestStruct{
		Name:  "John",
		Age:   30,
		Email: "john@example.com",
	}

	if result != expected {
		t.Errorf("MapToStruct() = %v, want %v", result, expected)
	}
}
