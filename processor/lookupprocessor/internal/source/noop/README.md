# noop Source

A no-operation source that always returns "not found". Useful for testing and benchmarking the processor overhead.

## Configuration

No configuration options.

## Example

```yaml
processors:
  lookup:
    source:
      type: noop
    attributes:
      - key: result
        from_attribute: key
        default: "not-found"
```
