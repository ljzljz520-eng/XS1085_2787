# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
--- FAIL: TestCloseWorkflowReturnsToQuietState (0.09s)
    integration_test.go:135: closed candle generated sparks: before=2 after=7
FAIL
FAIL	memorialcandle	0.122s
?   	memorialcandle/cmd/memorial	[no test files]
ok  	memorialcandle/internal/audit	0.003s
ok  	memorialcandle/internal/catalog	0.005s
?   	memorialcandle/internal/config	[no test files]
ok  	memorialcandle/internal/insights	0.004s
--- FAIL: TestCandleStopsAfterClose (0.14s)
    animation_test.go:19: spark stream continued after close: before=3 after=11
FAIL
FAIL	memorialcandle/internal/memorial	0.174s
ok  	memorialcandle/internal/protocol	0.008s
ok  	memorialcandle/internal/ritual	0.002s
ok  	memorialcandle/internal/store	0.059s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/memorial): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/memorial): exit `0`
