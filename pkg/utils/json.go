package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// PrettyJSON formats JSON with proper indentation
func PrettyJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CompactJSON removes whitespace from JSON
func CompactJSON(jsonStr string) (string, error) {
	var buffer bytes.Buffer
	err := json.Compact(&buffer, []byte(jsonStr))
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// JSONEqual compares two JSON strings for equality
func JSONEqual(json1, json2 string) bool {
	var obj1, obj2 interface{}
	
	if err := json.Unmarshal([]byte(json1), &obj1); err != nil {
		return false
	}
	
	if err := json.Unmarshal([]byte(json2), &obj2); err != nil {
		return false
	}
	
	return reflect.DeepEqual(obj1, obj2)
}

// ExtractJSONField extracts a specific field from JSON string
func ExtractJSONField(jsonStr, fieldPath string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}
	
	return getNestedField(data, strings.Split(fieldPath, "."))
}

// getNestedField recursively extracts nested field
func getNestedField(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}
	
	switch v := data.(type) {
	case map[string]interface{}:
		if value, exists := v[path[0]]; exists {
			return getNestedField(value, path[1:])
		}
		return nil, fmt.Errorf("field %s not found", path[0])
	default:
		return nil, fmt.Errorf("cannot access field %s on non-object", path[0])
	}
}

// MergeJSON merges multiple JSON objects
func MergeJSON(jsons ...string) (string, error) {
	result := make(map[string]interface{})
	
	for _, jsonStr := range jsons {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			return "", err
		}
		
		for key, value := range data {
			result[key] = value
		}
	}
	
	bytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	
	return string(bytes), nil
}

// FilterJSONFields filters JSON to include only specified fields
func FilterJSONFields(jsonStr string, fields []string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	
	result := make(map[string]interface{})
	for _, field := range fields {
		if value, exists := data[field]; exists {
			result[field] = value
		}
	}
	
	bytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	
	return string(bytes), nil
}

// ExcludeJSONFields filters JSON to exclude specified fields
func ExcludeJSONFields(jsonStr string, fields []string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	
	fieldSet := make(map[string]bool)
	for _, field := range fields {
		fieldSet[field] = true
	}
	
	result := make(map[string]interface{})
	for key, value := range data {
		if !fieldSet[key] {
			result[key] = value
		}
	}
	
	bytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	
	return string(bytes), nil
}

// JSONToQuery converts JSON object to URL query string
func JSONToQuery(jsonStr string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	
	var params []string
	for key, value := range data {
		params = append(params, fmt.Sprintf("%s=%v", key, value))
	}
	
	return strings.Join(params, "&"), nil
}

// QueryToJSON converts URL query string to JSON
func QueryToJSON(queryStr string) (string, error) {
	result := make(map[string]string)
	
	if queryStr == "" {
		bytes, _ := json.Marshal(result)
		return string(bytes), nil
	}
	
	pairs := strings.Split(queryStr, "&")
	for _, pair := range pairs {
		if eq := strings.Index(pair, "="); eq != -1 {
			key := pair[:eq]
			value := pair[eq+1:]
			result[key] = value
		}
	}
	
	bytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	
	return string(bytes), nil
}