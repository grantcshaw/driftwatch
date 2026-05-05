# driftwatch

Monitors infrastructure state and alerts on configuration drift between environments.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git && cd driftwatch && go build ./...
```

---

## Usage

Define your environments in a config file and run driftwatch to compare state:

```yaml
# driftwatch.yaml
environments:
  - name: staging
    provider: aws
    region: us-east-1
  - name: production
    provider: aws
    region: us-west-2
```

```bash
# Run a drift check
driftwatch check --config driftwatch.yaml

# Watch continuously and alert on drift
driftwatch watch --config driftwatch.yaml --interval 5m --alert slack
```

Example output:

```
[DRIFT DETECTED] staging vs production
  - instance_type: t3.medium vs t3.large
  - auto_scaling_min: 2 vs 4
No drift detected in: security_groups, vpc_settings
```

---

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `driftwatch.yaml` |
| `--interval` | Watch interval | `10m` |
| `--alert` | Alert destination (`slack`, `pagerduty`) | none |

---

## License

MIT © 2024 driftwatch contributors