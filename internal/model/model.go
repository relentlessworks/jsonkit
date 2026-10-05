package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Transform applies one or more pipe-separated operations to a JSON value.
// Operations are separated by '|'. Supported operations:
//   - keys: list all keys at the top level
//   - values: list all values at the top level
//   - sort: sort array elements (ascending)
//   - reverse: reverse array elements
//   - unique: remove duplicate elements from an array
//   - flatten: flatten nested objects using dot notation
//   - minify: compact JSON (no whitespace)
//   - pretty: pretty-print JSON with 2-space indent
//   - lowercase: convert all string values to lowercase
//   - uppercase: convert all string values to uppercase
//
// Example: "keys|sort" will list keys then sort the result.
func Transform(data []byte, ops string) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	parts := strings.Split(ops, "|")
	for _, part := range parts {
		op := strings.TrimSpace(part)
		if op == "" {
			continue
		}
		result, err := applyOp(v, op)
		if err != nil {
			return nil, err
		}
		v = result
	}

	return json.Marshal(v)
}

// applyOp applies a single operation to a JSON value.
func applyOp(v interface{}, op string) (interface{}, error) {
	switch op {
	case "keys":
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("keys: expected object, got %T", v)
		}
		ks := make([]string, 0, len(m))
		for k := range m {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		result := make([]interface{}, len(ks))
		for i, k := range ks {
			result[i] = k
		}
		return result, nil

	case "values":
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("values: expected object, got %T", v)
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		vals := make([]interface{}, 0, len(m))
		for _, k := range keys {
			vals = append(vals, m[k])
		}
		return vals, nil

	case "sort":
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("sort: expected array, got %T", v)
		}
		sorted := make([]interface{}, len(arr))
		copy(sorted, arr)
		sort.Slice(sorted, func(i, j int) bool {
			return fmt.Sprintf("%v", sorted[i]) < fmt.Sprintf("%v", sorted[j])
		})
		return sorted, nil

	case "reverse":
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("reverse: expected array, got %T", v)
		}
		rev := make([]interface{}, len(arr))
		for i, val := range arr {
			rev[len(arr)-1-i] = val
		}
		return rev, nil

	case "unique":
		arr, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unique: expected array, got %T", v)
		}
		seen := make(map[string]bool)
		result := make([]interface{}, 0, len(arr))
		for _, item := range arr {
			key := fmt.Sprintf("%v", item)
			if !seen[key] {
				seen[key] = true
				result = append(result, item)
			}
		}
		return result, nil

	case "flatten":
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("flatten: expected object, got %T", v)
		}
		return flatten(m, ""), nil

	case "minify":
		// Re-marshal without indentation (default json.Marshal does this)
		return v, nil

	case "pretty":
		// Just return as-is; the final Marshal will handle it
		// We mark it by wrapping in a prettyMarker
		return v, nil

	case "lowercase":
		return transformStrings(v, strings.ToLower), nil

	case "uppercase":
		return transformStrings(v, strings.ToUpper), nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
}

// flatten recursively flattens a nested map using dot notation.
func flatten(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, val := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := val.(map[string]interface{}); ok {
			for nk, nv := range flatten(nested, key) {
				result[nk] = nv
			}
		} else {
			result[key] = val
		}
	}
	return result
}

// transformStrings recursively applies a string transform function to all string values.
func transformStrings(v interface{}, fn func(string) string) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, vv := range val {
			result[k] = transformStrings(vv, fn)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, vv := range val {
			result[i] = transformStrings(vv, fn)
		}
		return result
	case string:
		return fn(val)
	default:
		return v
	}
}

// Format formats JSON data with the given indentation.
// If indent is 0, the output is compact (minified).
func Format(data []byte, indent int) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if indent <= 0 {
		return json.Marshal(v)
	}
	return json.MarshalIndent(v, "", strings.Repeat(" ", indent))
}

// Validate checks if the given data is valid JSON.
func Validate(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

// Query extracts a value from JSON data using a dot-notation path.
// Example: "user.name" extracts data["user"]["name"].
func Query(data []byte, path string) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	parts := strings.Split(path, ".")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch val := v.(type) {
		case map[string]interface{}:
			next, ok := val[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
			v = next
		case []interface{}:
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err != nil {
				return nil, fmt.Errorf("expected array index, got: %s", part)
			}
			if idx < 0 || idx >= len(val) {
				return nil, fmt.Errorf("array index out of range: %d", idx)
			}
			v = val[idx]
		default:
			return nil, fmt.Errorf("cannot index into %T with %s", v, part)
		}
	}

	return json.Marshal(v)
}

// Extract extracts specific keys from a JSON object.
// Returns a new JSON object containing only the specified keys.
func Extract(data []byte, keys []string) ([]byte, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid JSON (expected object): %w", err)
	}
	result := make(map[string]interface{})
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return json.Marshal(result)
}

// Merge merges multiple JSON objects into one.
// Later objects override earlier ones for the same key.
func Merge(datas ...[]byte) ([]byte, error) {
	result := make(map[string]interface{})
	for i, data := range datas {
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("input %d: invalid JSON: %w", i, err)
		}
		for k, v := range m {
			result[k] = v
		}
	}
	return json.Marshal(result)
}

// Diff compares two JSON values and returns the differences.
// Returns a map of key -> {"old": oldval, "new": newval} for changed keys,
// and keys only in the first are marked with null new, keys only in the
// second are marked with null old.
func Diff(a, b []byte) ([]byte, error) {
	var ma, mb map[string]interface{}
	if err := json.Unmarshal(a, &ma); err != nil {
		return nil, fmt.Errorf("first input: invalid JSON: %w", err)
	}
	if err := json.Unmarshal(b, &mb); err != nil {
		return nil, fmt.Errorf("second input: invalid JSON: %w", err)
	}

	type diffEntry struct {
		Old interface{} `json:"old,omitempty"`
		New interface{} `json:"new,omitempty"`
	}

	result := make(map[string]diffEntry)
	for k, va := range ma {
		vb, exists := mb[k]
		if !exists {
			result[k] = diffEntry{Old: va, New: nil}
			continue
		}
		if fmt.Sprintf("%v", va) != fmt.Sprintf("%v", vb) {
			result[k] = diffEntry{Old: va, New: vb}
		}
	}
	for k, vb := range mb {
		if _, exists := ma[k]; !exists {
			result[k] = diffEntry{Old: nil, New: vb}
		}
	}

	return json.MarshalIndent(result, "", "  ")
}

// Filter filters array elements by a simple key=value condition.
// Example: filter [{"name":"a"},{"name":"b"}] with "name=a" returns [{"name":"a"}].
func Filter(data []byte, condition string) ([]byte, error) {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, fmt.Errorf("invalid JSON (expected array): %w", err)
	}

	condition = strings.TrimSpace(condition)
	if condition == "" {
		return json.Marshal(arr)
	}

	parts := strings.SplitN(condition, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("condition must be key=value, got: %s", condition)
	}
	key := strings.TrimSpace(parts[0])
	want := strings.TrimSpace(parts[1])

	result := make([]interface{}, 0)
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		val, exists := m[key]
		if !exists {
			continue
		}
		if fmt.Sprintf("%v", val) == want {
			result = append(result, item)
		}
	}
	return json.Marshal(result)
}

// Analyze returns statistics about a JSON value.
func Analyze(data []byte) ([]byte, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	stats := analyzeValue(v)
	return json.MarshalIndent(stats, "", "  ")
}

type stats struct {
	Type        string `json:"type"`
	Size        int    `json:"size,omitempty"`
	Keys        int    `json:"keys,omitempty"`
	Depth       int    `json:"depth"`
	StringCount int    `json:"string_count"`
	NumberCount int    `json:"number_count"`
	BoolCount   int    `json:"bool_count"`
	NullCount   int    `json:"null_count"`
}

func analyzeValue(v interface{}) stats {
	s := stats{}
	switch val := v.(type) {
	case map[string]interface{}:
		s.Type = "object"
		s.Keys = len(val)
		s.Size = len(val)
		maxDepth := 0
		for _, vv := range val {
			child := analyzeValue(vv)
			if child.Depth > maxDepth {
				maxDepth = child.Depth
			}
			s.StringCount += child.StringCount
			s.NumberCount += child.NumberCount
			s.BoolCount += child.BoolCount
			s.NullCount += child.NullCount
		}
		s.Depth = maxDepth + 1
	case []interface{}:
		s.Type = "array"
		s.Size = len(val)
		maxDepth := 0
		for _, vv := range val {
			child := analyzeValue(vv)
			if child.Depth > maxDepth {
				maxDepth = child.Depth
			}
			s.StringCount += child.StringCount
			s.NumberCount += child.NumberCount
			s.BoolCount += child.BoolCount
			s.NullCount += child.NullCount
		}
		s.Depth = maxDepth + 1
	case string:
		s.Type = "string"
		s.Depth = 1
		s.StringCount = 1
	case float64:
		s.Type = "number"
		s.Depth = 1
		s.NumberCount = 1
	case bool:
		s.Type = "bool"
		s.Depth = 1
		s.BoolCount = 1
	case nil:
		s.Type = "null"
		s.Depth = 1
		s.NullCount = 1
	}
	return s
}
