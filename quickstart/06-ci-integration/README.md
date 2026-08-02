# Step 6 — CI Integration

Run reqmd in CI for validation gates and structured reports.

## JSON validation report

```sh
reqmd check --json 02-trace-your-spec/
```

Produces a structured JSON report with per-requirement pass/fail and trace
check results:

```json
{
  "version": 1,
  "exit_code": 0,
  "summary": {
    "total": 5,
    "valid": 5,
    "invalid": 0,
    "parse_errors": 0,
    "warnings": 1
  },
  "documents": [...],
  "parse_errors": []
}
```

The `exit_code` field mirrors the process exit code: `0` (all valid), `1`
(validation errors), `2` (parse error).

## GitHub Actions example

```yaml
# .github/workflows/validate.yml
name: Validate requirements
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go build -o reqmd ./cmd/reqmd
      - run: ./reqmd check --json requirements/ > report.json
      - run: ./reqmd check requirements/ --results ci-artifacts/
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: reqmd-report
          path: report.json
```

## Exit codes for CI gating

| Code | Meaning | CI action |
|------|---------|-----------|
| 0 | All valid | ✅ Pass |
| 1 | Validation errors | ❌ Fail |
| 2 | Parse error | ❌ Fail |

Only ERROR-level checks affect the exit code. WARNINGs are informational.

## JSON list and stats

```sh
reqmd ls --json 02-trace-your-spec/      # all requirements with attributes
reqmd stats --json 02-trace-your-spec/  # attribute-value breakdown per doc
```

## Check suppression in CI

Suppress per-requirement checks that are known/accepted:

```yaml
reqmd-suppress:
  - untraced
  - version-pin
```

## What's next

You can integrate reqmd into CI. Now pin the trace links so upstream changes
don't silently invalidate the chain — go to [Step 7](../07-version-pins/).