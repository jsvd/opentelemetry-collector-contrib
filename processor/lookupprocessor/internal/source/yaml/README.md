# yaml Source

Loads key-value mappings from a YAML file. The file should contain a flat map of string keys to values.

## Configuration

| Field | Description | Default |
| ----- | ----------- | ------- |
| `path` | Path to the YAML file (required) | - |

## Example

```yaml
processors:
  lookup:
    source:
      type: yaml
      path: /etc/otel/mappings.yaml
    attributes:
      - key: service.display_name
        from_attribute: service.name
```

Example mappings file (`mappings.yaml`):

```yaml
svc-frontend: "Frontend Web App"
svc-backend: "Backend API Service"
svc-worker: "Background Worker"
```

## Benchmarks

Measures only the source lookup operation (map access), isolated from processor overhead:

| Map Size | ns/op | allocs/op |
|----------|-------|-----------|
| 10 entries | 1,239 | 0 |
| 100 entries | 1,367 | 0 |
| 1,000 entries | 1,345 | 0 |
| 10,000 entries | 1,324 | 0 |

## TODO

- [ ] Live reload of YAML file
