package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/jsonkit/internal/model"
)

// Server holds the HTTP mux and dependencies.
type Server struct {
	mux *http.ServeMux
}

// New creates a new API server.
func New() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/help", s.help)
	s.mux.HandleFunc("/.well-known/agent.md", s.help)
	s.mux.HandleFunc("/format", s.format)
	s.mux.HandleFunc("/minify", s.minify)
	s.mux.HandleFunc("/validate", s.validate)
	s.mux.HandleFunc("/query", s.query)
	s.mux.HandleFunc("/extract", s.extract)
	s.mux.HandleFunc("/flatten", s.flatten)
	s.mux.HandleFunc("/unflatten", s.unflatten)
	s.mux.HandleFunc("/merge", s.merge)
	s.mux.HandleFunc("/diff", s.diff)
	s.mux.HandleFunc("/sort", s.sort)
	s.mux.HandleFunc("/filter", s.filter)
	s.mux.HandleFunc("/transform", s.transform)
	s.mux.HandleFunc("/analyze", s.analyze)
}

// wantsJSON checks if the client wants JSON output.
func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	return r.Header.Get("Accept") == "application/json"
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

// writeText writes a plain text response.
func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintln(w, msg)
}

// writeError writes an error response in the appropriate format.
func writeError(w http.ResponseWriter, r *http.Request, status int, msg, hint string) {
	if wantsJSON(r) {
		writeJSON(w, status, map[string]string{
			"error": msg,
			"hint":  hint,
		})
		return
	}
	writeText(w, status, fmt.Sprintf("error: %s | hint: %s", msg, hint))
}

// readBody reads the request body up to a reasonable limit.
func readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024)) // 10MB max
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty request body")
	}
	return body, nil
}

func (s *Server) help(w http.ResponseWriter, r *http.Request) {
	manual := `jsonkit - Agentic-first JSON manipulation service

ENDPOINTS:
  POST /format       Pretty-print JSON. Body=JSON, ?indent=N (default 2)
  POST /minify       Compact JSON (no whitespace). Body=JSON
  POST /validate     Check if JSON is valid. Body=JSON
  POST /query        Extract value by dot-path. Body=JSON, ?path=user.name
  POST /extract      Extract specific keys. Body=JSON, ?keys=name,age
  POST /flatten      Flatten nested objects. Body=JSON
  POST /unflatten    Unflatten dot-keys into nested objects. Body=JSON
  POST /merge        Merge multiple JSON objects. Body=JSON array of objects
  POST /diff         Diff two JSON objects. Body=JSON array [a, b]
  POST /sort         Sort array elements. Body=JSON array
  POST /filter       Filter array by key=value. Body=JSON array, ?cond=key=value
  POST /transform    Apply pipe-separated operations. Body=JSON, ?ops=keys|sort
  POST /analyze      Get JSON statistics. Body=JSON

TRANSFORM OPERATIONS (pipe-separated with |):
  keys, values, sort, reverse, unique, flatten, minify, pretty, lowercase, uppercase

EXAMPLES:
  curl -X POST localhost:8484/format -d '{"a":1}'
  curl -X POST localhost:8484/validate -d '{"a":1}'
  curl -X POST "localhost:8484/query?path=user.name" -d '{"user":{"name":"alice"}}'
  curl -X POST "localhost:8484/extract?keys=name,age" -d '{"name":"a","age":30,"x":1}'
  curl -X POST "localhost:8484/transform?ops=keys|sort" -d '{"c":3,"a":1,"b":2}'
  curl -X POST "localhost:8484/filter?cond=type=x" -d '[{"type":"x"},{"type":"y"}]'

ERRORS:
  All 4xx responses include: error: message | hint: what to do next

FORMAT:
  Plain text by default. Add ?format=json or Accept: application/json for JSON.`
	writeText(w, http.StatusOK, manual)
}

func (s *Server) format(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON document as the request body")
		return
	}
	indent := 2
	if v := r.URL.Query().Get("indent"); v != "" {
		fmt.Sscanf(v, "%d", &indent)
	}
	out, err := model.Format(body, indent)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is valid JSON")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) minify(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON document as the request body")
		return
	}
	out, err := model.Format(body, 0)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is valid JSON")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON document as the request body")
		return
	}
	if err := model.Validate(body); err != nil {
		if wantsJSON(r) {
			writeJSON(w, http.StatusOK, map[string]interface{}{"valid": false, "error": err.Error()})
		} else {
			writeText(w, http.StatusOK, "valid=false error="+err.Error())
		}
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, map[string]bool{"valid": true})
	} else {
		writeText(w, http.StatusOK, "valid=true")
	}
}

func (s *Server) query(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, r, http.StatusBadRequest, "missing path parameter", "add ?path=key.subkey to specify the dot-notation path to extract")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON document as the request body")
		return
	}
	out, err := model.Query(body, path)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "check that the path exists in the JSON document")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) extract(w http.ResponseWriter, r *http.Request) {
	keysParam := r.URL.Query().Get("keys")
	if keysParam == "" {
		writeError(w, r, http.StatusBadRequest, "missing keys parameter", "add ?keys=name,age to specify which keys to extract")
		return
	}
	keys := strings.Split(keysParam, ",")
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON object as the request body")
		return
	}
	out, err := model.Extract(body, keys)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is a valid JSON object")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) flatten(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON object as the request body")
		return
	}
	out, err := model.Transform(body, "flatten")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is a valid JSON object")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) unflatten(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON object with dot-separated keys as the request body")
		return
	}
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON: "+err.Error(), "ensure the request body is a valid JSON object")
		return
	}
	result := unflatten(m)
	out, _ := json.MarshalIndent(result, "", "  ")
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func unflatten(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		parts := strings.Split(k, ".")
		if len(parts) == 1 {
			result[k] = v
			continue
		}
		current := result
		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = v
			} else {
				if _, ok := current[part]; !ok {
					current[part] = make(map[string]interface{})
				}
				if next, ok := current[part].(map[string]interface{}); ok {
					current = next
				} else {
					newMap := make(map[string]interface{})
					current[part] = newMap
					current = newMap
				}
			}
		}
	}
	return result
}

func (s *Server) merge(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON array of objects to merge, e.g. [{\"a\":1},{\"b\":2}]")
		return
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(body, &arr); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON: "+err.Error(), "send a JSON array of objects to merge, e.g. [{\"a\":1},{\"b\":2}]")
		return
	}
	datas := make([][]byte, len(arr))
	for i, raw := range arr {
		datas[i] = raw
	}
	out, err := model.Merge(datas...)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure all array elements are valid JSON objects")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) diff(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON array of two objects to diff, e.g. [{\"a\":1},{\"a\":2}]")
		return
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(body, &arr); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON: "+err.Error(), "send a JSON array of two objects to diff, e.g. [{\"a\":1},{\"a\":2}]")
		return
	}
	if len(arr) != 2 {
		writeError(w, r, http.StatusBadRequest, fmt.Sprintf("expected 2 objects, got %d", len(arr)), "send exactly two JSON objects in an array, e.g. [{\"a\":1},{\"a\":2}]")
		return
	}
	out, err := model.Diff(arr[0], arr[1])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure both array elements are valid JSON objects")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) sort(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON array to sort, e.g. [3,1,2]")
		return
	}
	out, err := model.Transform(body, "sort")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is a valid JSON array")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) filter(w http.ResponseWriter, r *http.Request) {
	cond := r.URL.Query().Get("cond")
	if cond == "" {
		writeError(w, r, http.StatusBadRequest, "missing cond parameter", "add ?cond=key=value to filter array elements")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON array to filter, e.g. [{\"type\":\"x\"}]")
		return
	}
	out, err := model.Filter(body, cond)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is a valid JSON array and cond is key=value")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) transform(w http.ResponseWriter, r *http.Request) {
	ops := r.URL.Query().Get("ops")
	if ops == "" {
		writeError(w, r, http.StatusBadRequest, "missing ops parameter", "add ?ops=keys|sort to specify pipe-separated operations to apply")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON document as the request body")
		return
	}
	out, err := model.Transform(body, ops)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "check that ops are valid: keys, values, sort, reverse, unique, flatten, minify, pretty, lowercase, uppercase (separated by |)")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}

func (s *Server) analyze(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "could not read request body", "send a JSON document as the request body")
		return
	}
	out, err := model.Analyze(body)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error(), "ensure the request body is valid JSON")
		return
	}
	if wantsJSON(r) {
		writeJSON(w, http.StatusOK, json.RawMessage(out))
	} else {
		writeText(w, http.StatusOK, string(out))
	}
}
