package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ToString converts any value to string
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	default:
		// Try JSON encoding for complex types
		if bytes, err := json.Marshal(v); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// ToInt converts value to int
func ToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		// Check for overflow on 32-bit systems
		if v > 1<<31-1 {
			return 0, fmt.Errorf("value %d overflows int", v)
		}
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// ToFloat64 converts value to float64
func ToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// ToBool converts value to bool
func ToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0, nil
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0, nil
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		switch v {
		case "true", "1", "yes", "on", "y", "t":
			return true, nil
		case "false", "0", "no", "off", "n", "f", "":
			return false, nil
		default:
			return strconv.ParseBool(v)
		}
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// ToStringSlice converts value to []string
func ToStringSlice(value interface{}) ([]string, error) {
	if value == nil {
		return []string{}, nil
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = ToString(item)
		}
		return result, nil
	case string:
		// Try to parse as JSON array first
		var jsonArray []string
		if err := json.Unmarshal([]byte(v), &jsonArray); err == nil {
			return jsonArray, nil
		}
		// Fall back to comma-separated values
		if v == "" {
			return []string{}, nil
		}
		return strings.Split(v, ","), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []string", value)
	}
}

// ToIntSlice converts value to []int
func ToIntSlice(value interface{}) ([]int, error) {
	if value == nil {
		return []int{}, nil
	}

	switch v := value.(type) {
	case []int:
		return v, nil
	case []interface{}:
		result := make([]int, len(v))
		for i, item := range v {
			intVal, err := ToInt(item)
			if err != nil {
				return nil, err
			}
			result[i] = intVal
		}
		return result, nil
	case string:
		// Try to parse as JSON array first
		var jsonArray []int
		if err := json.Unmarshal([]byte(v), &jsonArray); err == nil {
			return jsonArray, nil
		}
		// Fall back to comma-separated values
		if v == "" {
			return []int{}, nil
		}
		parts := strings.Split(v, ",")
		result := make([]int, len(parts))
		for i, part := range parts {
			intVal, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, err
			}
			result[i] = intVal
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []int", value)
	}
}

// StructToMap converts struct to map[string]interface{}
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot convert nil to map")
	}

	result := make(map[string]interface{})

	// Use JSON marshal/unmarshal for simplicity
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	return result, err
}

// MapToStruct converts map to struct
func MapToStruct(m map[string]interface{}, result interface{}) error {
	if m == nil {
		return fmt.Errorf("cannot convert nil map to struct")
	}

	if result == nil {
		return fmt.Errorf("result interface cannot be nil")
	}

	// Use JSON marshal/unmarshal for simplicity
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}
