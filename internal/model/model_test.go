package model

import (
	"encoding/json"
	"testing"
)

func TestFormat(t *testing.T) {
	data := []byte(`{"name":"test","value":42}`)
	out, err := Format(data, 2)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	var v map[string]interface{}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if v["name"] != "test" {
		t.Errorf("expected name=test, got %v", v["name"])
	}
}

func TestFormatMinify(t *testing.T) {
	data := []byte(`{  "name"  :  "test"  ,  "value"  :  42  }`)
	out, err := Format(data, 0)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	expected := `{"name":"test","value":42}`
	if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{`{"key":"value"}`, false},
		{`[1,2,3]`, false},
		{`"hello"`, false},
		{`42`, false},
		{`{invalid}`, true},
		{``, true},
	}
	for _, tt := range tests {
		err := Validate([]byte(tt.input))
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestQuery(t *testing.T) {
	data := []byte(`{"user":{"name":"alice","age":30},"items":["a","b","c"]}`)
	tests := []struct {
		path    string
		want    string
		wantErr bool
	}{
		{"user.name", `"alice"`, false},
		{"user.age", `30`, false},
		{"items.0", `"a"`, false},
		{"items.2", `"c"`, false},
		{"user.email", ``, true},
		{"nonexistent", ``, true},
	}
	for _, tt := range tests {
		out, err := Query(data, tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("Query(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && string(out) != tt.want {
			t.Errorf("Query(%q) = %s, want %s", tt.path, string(out), tt.want)
		}
	}
}

func TestExtract(t *testing.T) {
	data := []byte(`{"name":"test","value":42,"extra":"remove","nested":{"a":1}}`)
	out, err := Extract(data, []string{"name", "value"})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if len(m) != 2 {
		t.Errorf("expected 2 keys, got %d", len(m))
	}
	if m["name"] != "test" {
		t.Errorf("expected name=test, got %v", m["name"])
	}
	if m["value"] != float64(42) {
		t.Errorf("expected value=42, got %v", m["value"])
	}
}

func TestMerge(t *testing.T) {
	a := []byte(`{"a":1,"b":2}`)
	b := []byte(`{"b":3,"c":4}`)
	out, err := Merge(a, b)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if m["a"] != float64(1) {
		t.Errorf("expected a=1, got %v", m["a"])
	}
	if m["b"] != float64(3) {
		t.Errorf("expected b=3 (overridden), got %v", m["b"])
	}
	if m["c"] != float64(4) {
		t.Errorf("expected c=4, got %v", m["c"])
	}
}

func TestDiff(t *testing.T) {
	a := []byte(`{"name":"alice","age":30,"city":"NYC"}`)
	b := []byte(`{"name":"alice","age":31,"country":"US"}`)
	out, err := Diff(a, b)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if _, ok := m["age"]; !ok {
		t.Error("expected age in diff")
	}
	if _, ok := m["city"]; !ok {
		t.Error("expected city in diff (removed)")
	}
	if _, ok := m["country"]; !ok {
		t.Error("expected country in diff (added)")
	}
	if _, ok := m["name"]; ok {
		t.Error("name should not be in diff (unchanged)")
	}
}

func TestFilter(t *testing.T) {
	data := []byte(`[{"name":"a","type":"x"},{"name":"b","type":"x"},{"name":"c","type":"y"}]`)
	out, err := Filter(data, "type=x")
	if err != nil {
		t.Fatalf("Filter failed: %v", err)
	}
	var arr []interface{}
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 results, got %d", len(arr))
	}
}

func TestTransformKeys(t *testing.T) {
	data := []byte(`{"c":3,"a":1,"b":2}`)
	out, err := Transform(data, "keys")
	if err != nil {
		t.Fatalf("Transform keys failed: %v", err)
	}
	var arr []string
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid array: %v", err)
	}
	if len(arr) != 3 || arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
		t.Errorf("expected [a,b,c], got %v", arr)
	}
}

func TestTransformSort(t *testing.T) {
	data := []byte(`[3,1,2]`)
	out, err := Transform(data, "sort")
	if err != nil {
		t.Fatalf("Transform sort failed: %v", err)
	}
	var arr []interface{}
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid array: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != float64(1) || arr[1] != float64(2) || arr[2] != float64(3) {
		t.Errorf("expected [1,2,3], got %v", arr)
	}
}

func TestTransformReverse(t *testing.T) {
	data := []byte(`[1,2,3]`)
	out, err := Transform(data, "reverse")
	if err != nil {
		t.Fatalf("Transform reverse failed: %v", err)
	}
	var arr []interface{}
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid array: %v", err)
	}
	if arr[0] != float64(3) || arr[1] != float64(2) || arr[2] != float64(1) {
		t.Errorf("expected [3,2,1], got %v", arr)
	}
}

func TestTransformUnique(t *testing.T) {
	data := []byte(`["a","b","a","c","b"]`)
	out, err := Transform(data, "unique")
	if err != nil {
		t.Fatalf("Transform unique failed: %v", err)
	}
	var arr []interface{}
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid array: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 unique elements, got %d", len(arr))
	}
}

func TestTransformFlatten(t *testing.T) {
	data := []byte(`{"a":{"b":{"c":1}}}`)
	out, err := Transform(data, "flatten")
	if err != nil {
		t.Fatalf("Transform flatten failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid object: %v", err)
	}
	if v, ok := m["a.b.c"]; !ok || v != float64(1) {
		t.Errorf("expected a.b.c=1, got %v", m)
	}
}

func TestTransformLowercase(t *testing.T) {
	data := []byte(`{"name":"HELLO","items":["A","B"]}`)
	out, err := Transform(data, "lowercase")
	if err != nil {
		t.Fatalf("Transform lowercase failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid object: %v", err)
	}
	if m["name"] != "hello" {
		t.Errorf("expected name=hello, got %v", m["name"])
	}
}

func TestTransformUppercase(t *testing.T) {
	data := []byte(`{"name":"hello"}`)
	out, err := Transform(data, "uppercase")
	if err != nil {
		t.Fatalf("Transform uppercase failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid object: %v", err)
	}
	if m["name"] != "HELLO" {
		t.Errorf("expected name=HELLO, got %v", m["name"])
	}
}

func TestTransformPipe(t *testing.T) {
	data := []byte(`{"c":3,"a":1,"b":2}`)
	out, err := Transform(data, "keys|sort")
	if err != nil {
		t.Fatalf("Transform keys|sort failed: %v", err)
	}
	var arr []string
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not valid array: %v", err)
	}
	if len(arr) != 3 || arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
		t.Errorf("expected [a,b,c], got %v", arr)
	}
}

func TestAnalyze(t *testing.T) {
	data := []byte(`{"name":"test","count":42,"active":true,"items":[1,2,3]}`)
	out, err := Analyze(data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if m["type"] != "object" {
		t.Errorf("expected type=object, got %v", m["type"])
	}
	if m["keys"] != float64(4) {
		t.Errorf("expected keys=4, got %v", m["keys"])
	}
}
