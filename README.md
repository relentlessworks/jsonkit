# jsonkit

Agentic-first JSON manipulation service. Format, minify, validate, query, extract, flatten, unflatten, merge, diff, sort, filter, transform, and analyze JSON data. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
# Build
make build

# Run
./jsonkit

# Or with custom port
./jsonkit -port 9000
```

## API Reference

All endpoints accept JSON in the request body and return plain text by default. Add `?format=json` or `Accept: application/json` for JSON output.

### POST /format
Pretty-print JSON with indentation.
```bash
curl -X POST localhost:8484/format -d '{"a":1,"b":2}'
curl -X POST "localhost:8484/format?indent=4" -d '{"a":1}'
```

### POST /minify
Compact JSON (no whitespace).
```bash
curl -X POST localhost:8484/minify -d '{  "a"  :  1  }'
```

### POST /validate
Check if JSON is valid.
```bash
curl -X POST localhost:8484/validate -d '{"a":1}'
# valid=true
```

### POST /query
Extract a value using dot-notation path.
```bash
curl -X POST "localhost:8484/query?path=user.name" -d '{"user":{"name":"alice"}}'
# "alice"
```

### POST /extract
Extract specific keys from an object.
```bash
curl -X POST "localhost:8484/extract?keys=name,age" -d '{"name":"a","age":30,"x":1}'
```

### POST /flatten
Flatten nested objects using dot notation.
```bash
curl -X POST localhost:8484/flatten -d '{"a":{"b":{"c":1}}}'
# {"a.b.c":1}
```

### POST /unflatten
Unflatten dot-separated keys into nested objects.
```bash
curl -X POST localhost:8484/unflatten -d '{"a.b.c":1}'
# {"a":{"b":{"c":1}}}
```

### POST /merge
Merge multiple JSON objects (later overrides earlier).
```bash
curl -X POST localhost:8484/merge -d '[{"a":1},{"b":2}]'
```

### POST /diff
Diff two JSON objects.
```bash
curl -X POST localhost:8484/diff -d '[{"a":1,"b":2},{"a":1,"c":3}]'
```

### POST /sort
Sort array elements (ascending).
```bash
curl -X POST localhost:8484/sort -d '[3,1,2]'
# [1,2,3]
```

### POST /filter
Filter array elements by key=value condition.
```bash
curl -X POST "localhost:8484/filter?cond=type=x" -d '[{"type":"x"},{"type":"y"}]'
```

### POST /transform
Apply one or more pipe-separated operations to JSON data. Operations are `|`-separated.

Supported operations: `keys`, `values`, `sort`, `reverse`, `unique`, `flatten`, `minify`, `pretty`, `lowercase`, `uppercase`

```bash
curl -X POST "localhost:8484/transform?ops=keys|sort" -d '{"c":3,"a":1,"b":2}'
# ["a","b","c"]

curl -X POST "localhost:8484/transform?ops=sort|reverse" -d '[1,2,3]'
# [3,2,1]

curl -X POST "localhost:8484/transform?ops=unique" -d '["a","b","a"]'
# ["a","b"]
```

### POST /analyze
Get statistics about a JSON document.
```bash
curl -X POST localhost:8484/analyze -d '{"name":"test","count":42}'
```

### GET /help
Returns a one-page operating manual for agents.
```bash
curl localhost:8484/help
```

## Configuration

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `-addr` | `JSONKIT_ADDR` | `0.0.0.0` | Listen address |
| `-port` | `JSONKIT_PORT` | `8484` | Listen port |
| `-secret` | `JSONKIT_SECRET` | (auto) | Token signing secret |

## Build

```bash
make build    # CGO_ENABLED=0, single binary
make test     # go test -race
make vet      # go vet
```

## License

MIT
