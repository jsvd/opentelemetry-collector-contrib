# dns Source

Performs reverse DNS lookups (PTR records) to resolve IP addresses to hostnames. Supports caching to minimize DNS queries.

## Configuration

| Field | Description | Default |
| ----- | ----------- | ------- |
| `record_type` | DNS record type (currently only `PTR` is supported) | `PTR` |
| `timeout` | Maximum time to wait for DNS query | `5s` |
| `server` | DNS server to use (e.g., `8.8.8.8:53`). Empty uses system resolver | - |
| `cache.enabled` | Enable caching | `true` |
| `cache.size` | Maximum cache entries | `10000` |
| `cache.ttl` | Time-to-live for cached entries | `5m` |
| `cache.negative_ttl` | TTL for "not found" entries | `1m` |

## Example

**PTR lookup (reverse DNS - IP to hostname):**

```yaml
processors:
  lookup:
    source:
      type: dns
      record_type: PTR
      cache:
        enabled: true
        size: 10000
        ttl: 5m
    attributes:
      - key: client.hostname
        from_attribute: client.ip
        default: "unknown"
```

## Benchmarks

Measures DNS lookup with and without caching (network latency varies):

| Scenario | ns/op | allocs/op |
|----------|-------|-----------|
| PTR lookup (no cache) | ~210,000 | 16 |
| PTR lookup (cached) | 39 | 0 |

**Cache speedup: ~5000x faster** for repeated lookups.

## TODO

- [ ] Support A record lookups (hostname to IPv4)
- [ ] Support AAAA record lookups (hostname to IPv6)
- [ ] Support TXT, CNAME, MX record lookups
